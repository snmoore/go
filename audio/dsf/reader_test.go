// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"io"
	"os"
	"testing"
)

// Table structure for a single reader test
type readerTest struct {
	// Description for the test
	description string
	// Name of the DSD stream file to read
	filename string
	// Is an error expected to be thrown?
	expectError bool
}

// Table of all reader tests
var readerTests = []readerTest{
	// Chunk order: should be DSD, fmt, data, metadata
	{"Reading a DSD stream file that has chunks out of order (fmt before DSD) should result in an error", "test/invalid_fmt_before_dsd.dsf", true},
	{"Reading a DSD stream file that has chunks out of order (data before DSD) should result in an error", "test/invalid_data_before_dsd.dsf", true},
	{"Reading a DSD stream file that has chunks out of order (data before fmt) should result in an error", "test/invalid_data_before_fmt.dsf", true},
	{"Reading a DSD stream file that has missing chunks (missing DSD) should result in an error", "test/invalid_missing_dsd.dsf", true},
	{"Reading a DSD stream file that has missing chunks (missing fmt) should result in an error", "test/invalid_missing_fmt.dsf", true},
	{"Reading a DSD stream file that has missing chunks (missing data) should result in an error", "test/invalid_missing_data.dsf", true},
	{"Reading a DSD stream file that has missing chunks (missing metadata) should result in an error", "test/invalid_missing_metadata.dsf", true},

	// Valid DSD stream file
	{"Reading a valid DSD stream file (without metadata) should not result in an error", "test/valid_without_metadata.dsf", false},
	{"Reading a valid DSD stream file (with metadata) should not result in an error", "test/valid_with_metadata.dsf", false},
}

// Run all tests
func TestReader(t *testing.T) {
	// Only log the chunk contents if verbose is enabled
	var logTo io.Writer
	if testing.Verbose() {
		logTo = os.Stdout
	}

	// Run each test
	for i, test := range readerTests {
		// Open the DSD stream file
		file, err := os.Open(test.filename)
		if err != nil {
			t.Errorf("FAIL Test %v: %v:\n%v", i, test.description, err.Error())
		}

		// Read and decode the DSD stream file
		_, err = Decode(file, logTo)

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

		// Close the DSD stream file
		if err := file.Close(); err != nil {
			t.Errorf("FAIL Test %v: %v:\n%v", i, test.description, err.Error())
		}
	}
}
