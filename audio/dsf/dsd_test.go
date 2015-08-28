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

// A valid DSD chunk
var validDsdChunk = []byte{
	// DSD chunk header: "DSD "
	'D', 'S', 'D', ' ',
	// Size of this chunk: 28 bytes
	0x1C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Total file size: at least 92 bytes for DSD, fmt and data chunks
	0x5C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Pointer to Metadata chunk: none present
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// Table driven DSD chunk tests
var dsdChunkTests = []test{
	// Chunk header: should be "DSD "
	{"Reading a DSD chunk that has an invalid chunk header (bad last byte) should result in an error", 0, []byte{'D', 'S', 'D', 'x'}, true},
	{"Reading a DSD chunk that has an invalid chunk header (lowercase) should result in an error", 0, []byte{'d', 's', 'd', ' '}, true},
	{"Reading a DSD chunk that has a valid chunk header should not result in an error", 0, []byte{'D', 'S', 'D', ' '}, false},
	{"Encountering a fmt chunk whilst reading a DSD chunk should result in an error", 0, []byte{'f', 'm', 't', ' '}, true},
	{"Encountering a data chunk whilst reading a DSD chunk should result in an error", 0, []byte{'d', 'a', 't', 'a'}, true},
	{"Encountering a metadata chunk whilst reading a DSD chunk should result in an error", 0, []byte{'I', 'D', '3', 0x03}, true},

	// Chunk size: should be 28 bytes
	{"Reading a DSD chunk that has an invalid chunk size (size - 1) should result in an error", 4, []byte{27}, true},
	{"Reading a DSD chunk that has an invalid chunk size (size + 1) should result in an error", 4, []byte{29}, true},
	{"Reading a DSD chunk that has a valid chunk size should not result in an error", 4, []byte{28}, false},

	// Total file size: should be at least 92 bytes for DSD, fmt and data chunks
	{"Reading a DSD chunk that has an invalid total file size (too small) should result in an error", 12, []byte{91}, true},
	{"Reading a DSD chunk that has a valid total file size should not result in an error", 12, []byte{92}, false},

	// Pointer to Metadata chunk: if present the metada should be located after the DSD, fmt and data chunks
	{"Reading a DSD chunk that has an invalid pointer to metadata (too small) should result in an error", 20, []byte{91}, true},
	{"Reading a DSD chunk that has an invalid pointer to metadata (too large) should result in an error", 20, []byte{92}, true},
	{"Reading a DSD chunk that has a valid pointer to metadata (none present) should not result in an error", 20, []byte{0}, false},
	{"Reading a DSD chunk that has a valid pointer to metadata (present) should not result in an error", 12, []byte{0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5D, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, false},

	// Sanity check: valid DSD chunk
	{"Reading a valid DSD chunk should not result in an error", 0, []byte{}, false},
}

// Run the table driven tests
func TestDsdRead(t *testing.T) {
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
	for i, test := range dsdChunkTests {
		// Start with a valid chunk
		c := make([]byte, len(validDsdChunk))
		copy(c, validDsdChunk)

		// Patch the test data into the valid chunk
		copy(c[test.offset:], test.data)

		// Read the chunk
		d.reader = bytes.NewReader(c)
		err := d.readDSDChunk()

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

// A read error whilst reading a DSD chunk should result in an error
func TestDsdReadError(t *testing.T) {
	description := "A read error whilst reading a DSD chunk should result in an error"

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
	err := d.readDSDChunk()

	// Reading the chunk should have thrown an error
	if err == nil {
		t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", len(dsdChunkTests)+1, description)
	} else {
		t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", len(dsdChunkTests)+1, description, err.Error())
	}
}
