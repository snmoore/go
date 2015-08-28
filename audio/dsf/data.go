// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"encoding/binary"
	"fmt"
)

// DataChunk is the file structure of the data chunk within a DSD stream file,
// excluding the variable length sample data. See "DSF File Format
// Specification", v1.01, Sony Corporation. All data is little-endian. This is
// exported to allow reading with binary.Read.
type DataChunk struct {
	// data chunk header.
	// 'd' , 'a' , 't', 'a '.
	Header [4]byte

	// Size of this chunk.
	Size [8]byte

	// Sample data, omitted because this is not a fixed size.
	// Samples []byte
}

// Header identifying a data chunk within a DSD stream file.
const dataChunkHeader = "data"

// Size in bytes of a data chunk within a DSD stream file, excluding samples.
const dataChunkSize = 12

// readDataChunk reads the data chunk and stores the result in d. The audio
// samples are typically huge (tens or hundreds of MB) and hence are written
// directly into the audio.Audio in d.
func (d *decoder) readDataChunk() error {
	// Read the chunk excluding the sample data
	err := binary.Read(d.reader, binary.LittleEndian, &d.data)
	if err != nil {
		return err
	}

	// Chunk header
	header := string(d.data.Header[:])
	switch header {
	case dataChunkHeader:
		// This is the expected chunk header
	case dsdChunkHeader:
		return fmt.Errorf("data: expected data chunk but found DSD chunk")
	case fmtChunkHeader:
		return fmt.Errorf("data: expected data chunk but found fmt chunk")
	default:
		return fmt.Errorf("data: bad chunk header: %q\ndata chunk: % x", header, d.data)
	}

	// Size of this chunk
	size := binary.LittleEndian.Uint64(d.data.Size[:])
	if size != dataChunkSize+uint64(len(d.audio.EncodedSamples)) {
		return fmt.Errorf("data: bad chunk size: %v\nfmt chunk: % x\ndata chunk: % x", size, d.fmt, d.data)
	}

	// Read the sample data directly into the audio.Audio in d
	err = binary.Read(d.reader, binary.LittleEndian, &d.audio.EncodedSamples)
	if err != nil {
		return err
	}

	// Log the fields of the chunk (only active if a log output has been set)
	d.logger.Print("\nData Chunk\n==========\n")
	d.logger.Printf("Chunk header:              %q\n", header)
	d.logger.Printf("Size of this chunk:        %v\n", size)
	if len(d.audio.EncodedSamples) > 0 {
		n := len(d.audio.EncodedSamples)
		if n > 20 {
			n = 20
		}
		d.logger.Printf("Sample data:               % x...\n", d.audio.EncodedSamples[:n])
	}

	return nil
}
