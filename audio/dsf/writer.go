// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"fmt"
	"github.com/snmoore/go/audio"
	"io"
	"log"
)

// encoder is the type used to encode a DSD stream file.
type encoder struct {
	// Where to log to.
	logger *log.Logger

	// Input.
	audio *audio.Audio

	// Output.
	writer io.Writer

	// DSD stream file chunks.
	dsd  DsdChunk
	fmt  FmtChunk
	data DataChunk
}

// encode writes a DSD stream file to r.
func (e *encoder) encode(a *audio.Audio, w io.Writer, logTo io.Writer) error {
	e.logger = log.New(logTo, "", 0)
	e.audio = a
	e.writer = w

	// Audio samples should be a multiple of the block size, padded with zero
	remainder := uint(len(e.audio.EncodedSamples)) % e.audio.BlockSize
	if remainder > 0 {
		e.logger.Printf("Padding the audio samples with %v zero bytes\n", remainder)
		padding := make([]byte, remainder, 0)
		e.audio.EncodedSamples = append(e.audio.EncodedSamples, padding...)
	}

	// Write the DSD stream file chunks
	if err := e.writeDSDChunk(); err != nil {
		return err
	}

	if err := e.writeFmtChunk(); err != nil {
		return err
	}

	return nil
}

// Encode writes the Audio a to w as a DSD stream file.
// logTo is the optional destination to log to.
func Encode(a *audio.Audio, w io.Writer, logTo io.Writer) error {
	var e encoder

	if a.Encoding != audio.DSD {
		return fmt.Errorf("unsupported audio encoding: %v\n", a.Encoding)
	}

	if err := e.encode(a, w, logTo); err != nil {
		return err
	}

	return nil
}
