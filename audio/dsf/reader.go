// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"github.com/snmoore/go/audio"
	"io"
	"io/ioutil"
	"log"
)

// decoder is the type used to decode a DSD stream file.
type decoder struct {
	// Where to log to.
	logger *log.Logger

	// Input.
	reader io.Reader

	// Output.
	audio *audio.Audio

	// DSD stream file chunks.
	dsd  DsdChunk
	fmt  FmtChunk
	data DataChunk
}

// decode reads a DSD stream file from r and stores the result in d.
func (d *decoder) decode(r io.Reader, logTo io.Writer) error {
	d.logger = log.New(logTo, "", 0)
	d.reader = r
	d.audio = new(audio.Audio)

	// 1st chunk should be DSD
	if err := d.readDSDChunk(); err != nil {
		return err
	}

	// 2nd chunk should be fmt
	if err := d.readFmtChunk(); err != nil {
		return err
	}

	// 3rd chunk should be data
	if err := d.readDataChunk(); err != nil {
		return err
	}

	// 4th chunk should be metadata, but may be omitted
	if len(d.audio.Metadata) > 0 {
		if err := d.readMetadataChunk(); err != nil {
			return err
		}
	}

	return nil
}

// Decode reads a DSD stream file from r and returns it as an Audio.
// logTo is the optional destination to log to.
func Decode(r io.Reader, logTo io.Writer) (*audio.Audio, error) {
	var d decoder

	if logTo == nil {
		logTo = ioutil.Discard
	}

	if err := d.decode(r, logTo); err != nil {
		return nil, err
	}

	return d.audio, nil
}
