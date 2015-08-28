// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"encoding/binary"
	"fmt"
)

// readMetadataChunk reads the metadata chunk and stores the result in d. This
// may be large and hence is written directly into the audio.Audio in d.
func (d *decoder) readMetadataChunk() error {
	// Read the metadata directly into the audio.Audio in d
	err := binary.Read(d.reader, binary.LittleEndian, &d.audio.Metadata)
	if err != nil {
		return err
	}

	// Check this is not just another DSD, fmt or data chunk
	header := string(d.audio.Metadata[:4])
	switch header {
	case dsdChunkHeader:
		return fmt.Errorf("metadata: expected metadata chunk but found DSD chunk")
	case fmtChunkHeader:
		return fmt.Errorf("metadata: expected metadata chunk but found fmt chunk")
	case dataChunkHeader:
		return fmt.Errorf("metadata: expected metadata chunk but found data chunk")
	default:
		// Anything else is acceptable
	}

	if len(d.audio.Metadata) > 0 {
		// Log the fields of the chunk (only active if a log output has been set)
		d.logger.Print("\nMetadata Chunk\n==============\n")
		d.logger.Printf("Size of metadata:          %v bytes\n", len(d.audio.Metadata))
		n := len(d.audio.Metadata)
		if n > 20 {
			n = 20
		}
		d.logger.Printf("Metadata:                  % x...\n", d.audio.Metadata[:n])
	}

	return nil
}
