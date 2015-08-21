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

// A valid data chunk
var validDataChunk = []byte{
	// data chunk header: "data"
	'd', 'a', 't', 'a',
	// Size of this chunk: 12 bytes plus 4096 bytes of sample data
	0x0C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Sample data: none present
}

// Table driven data chunk tests
var dataChunkTests = []test{
	// Chunk header: should be "data"
	{"Reading a data chunk that has an invalid chunk header (bad byte) should result in an error", 0, []byte{'s', 'a', 't', 'a'}, true},
	{"Reading a data chunk that has an invalid chunk header (uppercase) should result in an error", 0, []byte{'D', 'A', 'T', 'A'}, true},
	{"Reading a data chunk that has a valid chunk header should not result in an error", 0, []byte{'d', 'a', 't', 'a'}, false},
	{"Encountering a DSD chunk whilst reading a data chunk should result in an error", 0, []byte{'D', 'S', 'D', ' '}, true},
	{"Encountering a fmt chunk whilst reading a data chunk should result in an error", 0, []byte{'f', 'm', 't', ' '}, true},

	// Chunk size: should be 12 bytes plus the size of the sample data
	{"Reading a data chunk that has an invalid chunk size (too small) should result in an error", 4, []byte{11}, true},
	{"Reading a data chunk that has a valid chunk size should not result in an error", 4, []byte{12}, false},

	// Sample data: this needs special handing so is tested separately

	// Sanity check: valid DSD chunk
	{"Reading a valid data chunk should not result in an error", 0, []byte{}, false},
}

// Run the table driven tests
func TestData(t *testing.T) {
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
	for i, test := range dataChunkTests {
		// Start with a valid chunk
		c := make([]byte, len(validDataChunk))
		copy(c, validDataChunk)

		// Patch the test data into the valid chunk
		copy(c[test.offset:], test.data)

		// Read the chunk
		d.reader = bytes.NewReader(c)
		err := d.readDataChunk()

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

// Reading a data chunk that has an insufficient number of samples should result in an error
func TestDataSamplesInsufficient(t *testing.T) {
	description := "Reading a data chunk that has an insufficient number of samples should result in an error"

	// Prepare a decoder to use
	var d decoder
	d.audio = new(audio.Audio)

	// Only log the chunk contents if verbose is enabled
	if testing.Verbose() {
		d.logger = log.New(os.Stdout, "", 0)
	} else {
		d.logger = log.New(ioutil.Discard, "", 0)
	}

	// Start with a valid chunk
	c := make([]byte, len(validDataChunk))
	copy(c, validDataChunk)

	// Expect 4096 bytes of sample data, but do not actually provide them
	copy(c[4:], []byte{0x0C, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	d.audio.EncodedSamples = make([]byte, 4096)

	// Reading the chunk should throw an error
	d.reader = bytes.NewReader(c)
	err := d.readDataChunk()
	if err == nil {
		t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", len(dataChunkTests)+1, description)
	} else {
		t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", len(dataChunkTests)+1, description, err.Error())
	}
}

// Samples are read correctly from a data chunk
func TestDataSamplesRead(t *testing.T) {
	description := "Samples are read correctly from a data chunk"

	// Prepare 4096 bytes of sample data
	samples := make([]byte, 4096)
	for i, _ := range samples {
		samples[i] = byte(i)
	}

	// Prepare a decoder to use
	var d decoder
	d.audio = new(audio.Audio)

	// Only log the chunk contents if verbose is enabled
	if testing.Verbose() {
		d.logger = log.New(os.Stdout, "", 0)
	} else {
		d.logger = log.New(ioutil.Discard, "", 0)
	}

	// Start with a valid chunk
	c := make([]byte, len(validDataChunk))
	copy(c, validDataChunk)

	// Use the 4096 bytes of sample data prepared previously
	copy(c[4:], []byte{0x0C, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	c = append(c, samples...)

	// Prepare the decoder to expect 4096 bytes of sample data
	// This is normally done when reading a fmt chunk, omitted in this test
	d.audio.EncodedSamples = make([]byte, 4096)

	// Reading the chunk should not throw an error
	d.reader = bytes.NewReader(c)
	err := d.readDataChunk()
	if err != nil {
		t.Fatalf("FAIL Test %v: %v:\nWant: nil\nActual: %v", len(dataChunkTests)+2, description, err.Error())
	} else {
		t.Logf("PASS Test %v: %v:\nWant: nil\nActual: nil", len(dataChunkTests)+2, description)
	}

	// Verify the samples were read correctly
	for j, sample := range samples {
		if d.audio.EncodedSamples[j] != sample {
			t.Fatalf("FAIL Test %v: %v:\nIncorrect sample data at byte %v: %v != %v",
				len(dataChunkTests)+2, description, j, d.audio.EncodedSamples[j], sample)
		}
	}
}

// A read error whilst reading a data chunk should result in an error
func TestDataReadError(t *testing.T) {
	description := "A read error whilst reading a data chunk should result in an error"

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
	err := d.readDataChunk()

	// Reading the chunk should have thrown an error
	if err == nil {
		t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", len(dataChunkTests)+3, description)
	} else {
		t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", len(dataChunkTests)+3, description, err.Error())
	}
}
