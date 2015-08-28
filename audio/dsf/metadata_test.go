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

// A valid metadata chunk i.e. an ID3v2.3.0 tag
var validMetadataChunk = []byte{
	// File identifier: "ID3"
	'I', 'D', '3',
	// Major version: 3
	0x03,
	// Revision number: 0
	0x00,
	// Flags: none
	0x00,
	// Size: 0 bytes, excludes the tag header
	0x00, 0x00, 0x00, 0x00,
}

// Table driven metadata chunk tests
var metadataChunkTests = []test{
	{"Encountering a DSD chunk whilst reading a metadata chunk should result in an error", 0, []byte{'D', 'S', 'D', ' '}, true},
	{"Encountering a fmt chunk whilst reading a metadata chunk should result in an error", 0, []byte{'f', 'm', 't', ' '}, true},
	{"Encountering a data chunk whilst reading a metadata chunk should result in an error", 0, []byte{'d', 'a', 't', 'a'}, true},
}

// Run the table driven tests
func TestMetadataRead(t *testing.T) {
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
	for i, test := range metadataChunkTests {
		// Start with a valid chunk
		c := make([]byte, len(validMetadataChunk))
		copy(c, validMetadataChunk)

		// Patch the test data into the valid chunk
		copy(c[test.offset:], test.data)

		// Read the chunk
		d.audio.Metadata = make([]byte, len(validMetadataChunk))
		d.reader = bytes.NewReader(c)
		err := d.readMetadataChunk()

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

// A read error whilst reading a metadata chunk should result in an error
func TestMetadataReadError(t *testing.T) {
	description := "A read error whilst reading a metadata chunk should result in an error"

	// Prepare a decoder to use
	var d decoder
	d.audio = new(audio.Audio)

	// Only log the chunk contents if verbose is enabled
	if testing.Verbose() {
		d.logger = log.New(os.Stdout, "", 0)
	} else {
		d.logger = log.New(ioutil.Discard, "", 0)
	}

	// Prepare the decoder to expect 1024 bytes of metadata
	// This is normally done when reading a DSD chunk, omitted in this test
	d.audio.Metadata = make([]byte, 1024)

	// Read an empty chunk to force a read error
	d.reader = bytes.NewReader([]byte{})
	err := d.readMetadataChunk()

	// Reading the chunk should have thrown an error
	if err == nil {
		t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", len(metadataChunkTests)+1, description)
	} else {
		t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", len(metadataChunkTests)+1, description, err.Error())
	}
}

// Reading a metadata chunk that has an insufficient number of bytes should result in an error
func TestMetadataReadInsufficientBytes(t *testing.T) {
	description := "Reading a metadata chunk that has an insufficient number of bytes should result in an error"

	// Prepare a decoder to use
	var d decoder
	d.audio = new(audio.Audio)

	// Only log the chunk contents if verbose is enabled
	if testing.Verbose() {
		d.logger = log.New(os.Stdout, "", 0)
	} else {
		d.logger = log.New(ioutil.Discard, "", 0)
	}

	// Expect 1024 bytes of metadata, but do not actually provide them
	c := make([]byte, 0)
	d.audio.Metadata = make([]byte, 1024)

	// Reading the chunk should throw an error
	d.reader = bytes.NewReader(c)
	err := d.readDataChunk()
	if err == nil {
		t.Errorf("FAIL Test %v: %v:\nWant: error\nActual: nil", len(metadataChunkTests)+2, description)
	} else {
		t.Logf("PASS Test %v: %v:\nWant: error\nActual: %v", len(metadataChunkTests)+2, description, err.Error())
	}
}

// Bytes are read correctly from a metadata chunk
func TestMetadataReadBytes(t *testing.T) {
	description := "Samples are read correctly from a metadata chunk"

	// Prepare 1024 bytes of metadata
	metadata := make([]byte, 1024)
	for i, _ := range metadata {
		metadata[i] = byte(i)
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

	// Use the 1024 bytes of metadata prepared previously
	c := make([]byte, len(metadata))
	copy(c, metadata)

	// Prepare the decoder to expect 1024 bytes of metadata
	// This is normally done when reading a DSD chunk, omitted in this test
	d.audio.Metadata = make([]byte, 1024)

	// Reading the chunk should not throw an error
	d.reader = bytes.NewReader(c)
	err := d.readMetadataChunk()
	if err != nil {
		t.Fatalf("FAIL Test %v: %v:\nWant: nil\nActual: %v", len(metadataChunkTests)+3, description, err.Error())
	} else {
		t.Logf("PASS Test %v: %v:\nWant: nil\nActual: nil", len(metadataChunkTests)+3, description)
	}

	// Verify the bytes were read correctly
	for j, b := range metadata {
		if d.audio.Metadata[j] != b {
			t.Fatalf("FAIL Test 2: %v:\nIncorrect metadata at byte %v: %v != %v",
				description, j, d.audio.Metadata[j], b)
		}
	}
}
