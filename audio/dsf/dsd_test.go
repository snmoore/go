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

// Table of all DSD chunk tests
var dsdChunkTests = []test{
//	// Chunk header: should be "fmt "
//	{"Reading a DSD chunk that has an invalid chunk header (bad last byte) should result in an error", 0, []byte{'D', 'S', 'D', 'x'}, true},
//	{"Reading a DSD chunk that has an invalid chunk header (lowercase) should result in an error", 0, []byte{'d', 's', 'd', ' '}, true},
//	{"Reading a DSD chunk that has a valid chunk header should not result in an error", 0, []byte{'D', 'S', 'D', ' '}, false},
//
//	// Chunk size: should be 28 bytes
//	{"Reading a DSD chunk that has an invalid chunk size (size - 1) should result in an error", 4, []byte{27}, true},
//	{"Reading a DSD chunk that has an invalid chunk size (size + 1) should result in an error", 4, []byte{29}, true},
//	{"Reading a DSD chunk that has a valid chunk size should not result in an error", 4, []byte{28}, false},
//
//	// Total file size: should be at least 92 bytes for DSD, fmt and data chunks
//	{"Reading a DSD chunk that has an invalid total file size (too small) should result in an error", 12, []byte{91}, true},
//	{"Reading a DSD chunk that has a valid total file size should not result in an error", 12, []byte{92}, false},
//
//	// Pointer to Metadata chunk: if present the metada should be located after the DSD, fmt and data chunks
//	{"Reading a DSD chunk that has an invalid pointer to metadata (too small) should result in an error", 20, []byte{91}, true},
//	{"Reading a DSD chunk that has an invalid pointer to metadata (too large) should result in an error", 20, []byte{92}, true},
//	{"Reading a DSD chunk that has a valid pointer to metadata (none present) should not result in an error", 20, []byte{0}, false},
//	{"Reading a DSD chunk that has a valid pointer to metadata (present) should not result in an error", 12, []byte{0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5D, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, false},
//
//	// Sanity check: valid DSD chunk
//	{"Reading a valid DSD chunk should not result in an error", 0, []byte{}, false},
}

// Run all tests
func TestDsd(t *testing.T) {
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
		// Start with a valid fmt chunk
		c := make([]byte, len(validDsdChunk))
		copy(c, validDsdChunk)

		// Patch the test data into the valid fmt chunk
		copy(c[test.offset:], test.data)

		// Read the chunk
		d.reader = bytes.NewReader(c)
		err := d.readDSDChunk()

		// Check the result from reading the chunk
		if test.expectError {
			// Reading the chunk should have thrown an error
			if err == nil {
				t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", i, test.description)
			} else {
				t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", i, test.description, err.Error())
			}
		} else {
			// Reading the chunk should not have thrown an error
			if err != nil {
				t.Errorf("FAIL Test %v: %v:\nWant: nil\nActual: %v", i, test.description, err.Error())
			} else {
				t.Logf("PASS Test %v: %v:\nWant: nil\nActual: nil", i, test.description)
			}
		}
	}
}
