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

// Reading a metadata chunk that has an insufficient number of bytes should result in an error
func TestMetadataInsufficient(t *testing.T) {
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
		t.Errorf("FAIL Test 1: %v:\nWant: error\nActual: nil", description)
	} else {
		t.Logf("PASS Test 1: %v:\nWant: error\nActual: %v", description, err.Error())
	}
}

// Bytes are read correctly from a metadata chunk
func TestMetadataRead(t *testing.T) {
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
		t.Fatalf("FAIL Test 2: %v:\nWant: nil\nActual: %v", description, err.Error())
	} else {
		t.Logf("PASS Test 2: %v:\nWant: nil\nActual: nil", description)
	}

	// Verify the bytes were read correctly
	for j, b := range metadata {
		if d.audio.Metadata[j] != b {
			t.Fatalf("FAIL Test 2: %v:\nIncorrect metadata at byte %v: %v != %v",
				description, j, d.audio.Metadata[j], b)
		}
	}
}
