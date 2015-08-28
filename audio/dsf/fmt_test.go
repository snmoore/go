// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"bytes"
	"github.com/snmoore/go/audio"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// A valid fmt chunk
var validFmtChunk = []byte{
	// fmt chunk header: "fmt "
	'f', 'm', 't', ' ',
	// Size of this chunk: 52 bytes
	0x34, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Format version: 1
	0x01, 0x00, 0x00, 0x00,
	// Format id: 0 (DSD raw)
	0x00, 0x00, 0x00, 0x00,
	// Channel type: stereo
	0x02, 0x00, 0x00, 0x00,
	// Channel num: stereo
	0x02, 0x00, 0x00, 0x00,
	// Sampling frequency: 2822400 Hz
	0x00, 0x11, 0x2b, 0x00,
	// Bits per sample: 1
	0x01, 0x00, 0x00, 0x00,
	// Sample count: 1
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Block size per channel: 4096 bytes
	0x00, 0x10, 0x00, 0x00,
	// Reserved: filled with zero
	0x00, 0x00, 0x00, 0x00,
}

// Table structure for a single test
type test struct {
	// Description for the test
	description string
	// Byte offset into a valid chunk at which to patch the test data
	offset int
	// Test data to be patched into a valid chunk
	data []byte
	// Is an error expected to be thrown?
	expectError bool
}

// Table driven fmt chunk tests
var fmtChunkTests = []test{
	// Chunk header: should be "fmt "
	{"Reading a fmt chunk that has an invalid chunk header (bad last byte) should result in an error", 0, []byte{'f', 'm', 't', 'x'}, true},
	{"Reading a fmt chunk that has an invalid chunk header (uppercase) should result in an error", 0, []byte{'F', 'M', 'T', ' '}, true},
	{"Reading a fmt chunk that has a valid chunk header should not result in an error", 0, []byte{'f', 'm', 't', ' '}, false},
	{"Encountering a DSD chunk whilst reading a fmt chunk should result in an error", 0, []byte{'D', 'S', 'D', ' '}, true},
	{"Encountering a data chunk whilst reading a fmt chunk should result in an error", 0, []byte{'d', 'a', 't', 'a'}, true},
	{"Encountering a metadata chunk whilst reading a fmt chunk should result in an error", 0, []byte{'I', 'D', '3', 0x03}, true},

	// Chunk size: should be 52 bytes
	{"Reading a fmt chunk that has an invalid chunk size (size - 1) should result in an error", 4, []byte{51}, true},
	{"Reading a fmt chunk that has an invalid chunk size (size + 1) should result in an error", 4, []byte{53}, true},
	{"Reading a fmt chunk that has a valid chunk size should not result in an error", 4, []byte{52}, false},

	// Format version: should be 1
	{"Reading a fmt chunk that has an invalid format version (0) should result in an error", 12, []byte{0}, true},

	// Format id: should be 0 (DSD raw)
	{"Reading a fmt chunk that has an invalid format id (1) should result in an error", 16, []byte{1}, true},

	// Channel type: should be 1..7
	// This also tests channel num because the channel type and channel num have to match to avoid an error
	{"Reading a fmt chunk that has an invalid channel type (0) should result in an error", 20, []byte{0}, true},
	{"Reading a fmt chunk that has an invalid channel type (8) should result in an error", 20, []byte{8}, true},
	{"Reading a fmt chunk that has a valid channel type (mono) should not result in an error", 20, []byte{1, 0, 0, 0, 1, 0, 0, 0}, false},
	{"Reading a fmt chunk that has a valid channel type (stereo) should not result in an error", 20, []byte{2, 0, 0, 0, 2, 0, 0, 0}, false},
	{"Reading a fmt chunk that has a valid channel type (3 channels) should not result in an error", 20, []byte{3, 0, 0, 0, 3, 0, 0, 0}, false},
	{"Reading a fmt chunk that has a valid channel type (quad) should not result in an error", 20, []byte{4, 0, 0, 0, 4, 0, 0, 0}, false},
	{"Reading a fmt chunk that has a valid channel type (4 channels) should not result in an error", 20, []byte{5, 0, 0, 0, 4, 0, 0, 0}, false},
	{"Reading a fmt chunk that has a valid channel type (5 channels) should not result in an error", 20, []byte{6, 0, 0, 0, 5, 0, 0, 0}, false},
	{"Reading a fmt chunk that has a valid channel type (5.1 channels) should not result in an error", 20, []byte{7, 0, 0, 0, 6, 0, 0, 0}, false},

	// Channel num: should be 1..6
	// The valid numbers have already been tested in the channel type tests above
	{"Reading a fmt chunk that has an invalid number of channels (0) should result in an error", 24, []byte{0}, true},
	{"Reading a fmt chunk that has an invalid number of channels (7) should result in an error", 24, []byte{7}, true},

	// Channel type and channel num should match
	{"Reading a fmt chunk that has mismatched channel type and number of channels (mono) should result in an error", 20, []byte{1, 0, 0, 0, 2, 0, 0, 0}, true},
	{"Reading a fmt chunk that has mismatched channel type and number of channels (stereo) should result in an error", 20, []byte{2, 0, 0, 0, 1, 0, 0, 0}, true},
	{"Reading a fmt chunk that has mismatched channel type and number of channels (3 channels) should result in an error", 20, []byte{3, 0, 0, 0, 2, 0, 0, 0}, true},
	{"Reading a fmt chunk that has mismatched channel type and number of channels (4 channels) should result in an error", 20, []byte{4, 0, 0, 0, 3, 0, 0, 0}, true},
	{"Reading a fmt chunk that has mismatched channel type and number of channels (quad) should result in an error", 20, []byte{5, 0, 0, 0, 3, 0, 0, 0}, true},
	{"Reading a fmt chunk that has mismatched channel type and number of channels (5 channels) should result in an error", 20, []byte{6, 0, 0, 0, 4, 0, 0, 0}, true},
	{"Reading a fmt chunk that has mismatched channel type and number of channels (5.1 channels) should result in an error", 20, []byte{7, 0, 0, 0, 5, 0, 0, 0}, true},
	{"Reading a fmt chunk that has matched channel type and number of channels (mono) should not result in an error", 20, []byte{1, 0, 0, 0, 1, 0, 0, 0}, false},
	{"Reading a fmt chunk that has matched channel type and number of channels (stereo) should not result in an error", 20, []byte{2, 0, 0, 0, 2, 0, 0, 0}, false},
	{"Reading a fmt chunk that has matched channel type and number of channels (3 channels) should not result in an error", 20, []byte{3, 0, 0, 0, 3, 0, 0, 0}, false},
	{"Reading a fmt chunk that has matched channel type and number of channels (quad) should not result in an error", 20, []byte{4, 0, 0, 0, 4, 0, 0, 0}, false},
	{"Reading a fmt chunk that has matched channel type and number of channels (4 channels) should not result in an error", 20, []byte{5, 0, 0, 0, 4, 0, 0, 0}, false},
	{"Reading a fmt chunk that has matched channel type and number of channels (5 channels) should not result in an error", 20, []byte{6, 0, 0, 0, 5, 0, 0, 0}, false},
	{"Reading a fmt chunk that has matched channel type and number of channels (5.1 channels) should not result in an error", 20, []byte{7, 0, 0, 0, 6, 0, 0, 0}, false},

	// Sampling frequency: should be 2822400Hz, 5644800Hz, 11289600Hz or 22579200Hz
	// Only 2822400Hz and 5644800Hz are defined by the specification, but the other rates are in active use
	{"Reading a fmt chunk that has an invalid sampling frequency (44100Hz) should result in an error", 28, []byte{0x44, 0xAC, 0x00, 0x00}, true},
	{"Reading a fmt chunk that has a valid sampling frequency (2822400Hz) should not result in an error", 28, []byte{0x00, 0x11, 0x2B, 0x00}, false},
	{"Reading a fmt chunk that has a valid sampling frequency (5644800Hz) should not result in an error", 28, []byte{0x00, 0x22, 0x56, 0x00}, false},
	{"Reading a fmt chunk that has a valid sampling frequency (11289600Hz) should not result in an error", 28, []byte{0x00, 0x44, 0xAC, 0x00}, false},
	{"Reading a fmt chunk that has a valid sampling frequency (22579200Hz) should not result in an error", 28, []byte{0x00, 0x88, 0x58, 0x01}, false},

	// Bits per sample: should be 1 or 8
	{"Reading a fmt chunk that has an invalid number of bits per sample (0) should result in an error", 32, []byte{0}, true},
	{"Reading a fmt chunk that has a valid number of bits per sample (1) should not result in an error", 32, []byte{1}, false},
	{"Reading a fmt chunk that has a valid number of bits per sample (8) should not result in an error", 32, []byte{8}, false},

	// Sample count: no tests because any value is valid

	// Block size per channel: should be 4096 bytes
	{"Reading a fmt chunk that has an invalid block size (1024) should result in an error", 44, []byte{0x00, 0x04, 0x00, 0x00}, true},
	{"Reading a fmt chunk that has an valid block size (4096) should not result in an error", 44, []byte{0x00, 0x10, 0x00, 0x00}, false},

	// Reserved bytes: should be set to zero
	{"Reading a fmt chunk that has invalid reserved bytes (non-zero) should result in an error", 48, []byte{0x01, 0x02, 0x03, 0x04}, true},

	// Sanity check: valid fmt chunk
	{"Reading a valid fmt chunk should not result in an error", 0, []byte{}, false},
}

// Run the table driven tests
func TestFmtRead(t *testing.T) {
	// Prepare a decoder to use for all tests
	var d decoder
	d.audio = new(audio.Audio)

	// Only log the chunk contents if verbose is enabled
	if testing.Verbose() {
		d.logger = log.New(os.Stdout, "", 0)
	} else {
		d.logger = log.New(ioutil.Discard, "", 0)
	}

	// Run each test
	for i, test := range fmtChunkTests {
		// Start with a valid chunk
		c := make([]byte, len(validFmtChunk))
		copy(c, validFmtChunk)

		// Patch the test data into the valid chunk
		copy(c[test.offset:], test.data)

		// Read the chunk
		d.reader = bytes.NewReader(c)
		err := d.readFmtChunk()

		// Check the result from reading the chunk
		if test.expectError {
			// Reading the chunk should have thrown an error
			if err == nil {
				t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", i+1, test.description)
			} else {
				t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", i+1, test.description, err.Error())
			}
		} else {
			// Reading the chunk should not have thrown an error
			if err != nil {
				t.Errorf("FAIL Test %v: %v:\nWant: nil\nActual: %v", i+1, test.description, err.Error())
			} else {
				t.Logf("PASS Test %v: %v:\nWant: nil\nActual: nil", i+1, test.description)
			}
		}
	}
}

// A read error whilst reading a fmt chunk should result in an error
func TestFmtReadError(t *testing.T) {
	description := "A read error whilst reading a fmt chunk should result in an error"

	// Prepare a decoder to use
	var d decoder
	d.audio = new(audio.Audio)

	// Only log the chunk contents if verbose is enabled
	if testing.Verbose() {
		d.logger = log.New(os.Stdout, "", 0)
	} else {
		d.logger = log.New(ioutil.Discard, "", 0)
	}

	// Read an empty chunk to force a read error
	d.reader = bytes.NewReader([]byte{})
	err := d.readFmtChunk()

	// Reading the chunk should have thrown an error
	if err == nil {
		t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", len(fmtChunkTests)+1, description)
	} else {
		t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", len(fmtChunkTests)+1, description, err.Error())
	}
}
