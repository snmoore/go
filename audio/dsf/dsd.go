// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"encoding/binary"
	"fmt"
)

// DsdChunk is the file structure of the DSD chunk within a DSD stream file.
// See "DSF File Format Specification", v1.01, Sony Corporation. All data is
// little-endian. This is exported to allow reading with binary.Read.
type DsdChunk struct {
	// DSD chunk header.
	// 'D' , 'S' , 'D', ' ' (includes 1 space).
	Header [4]byte

	// Size of this chunk.
	// 28 bytes.
	Size [8]byte

	// Total file size.
	TotalFileSize [8]byte

	// Pointer to Metadata chunk.
	// If Metadata doesn’t exist, set 0. If the file has ID3v2 tag, then set the
	// pointer to it. ID3v2 tag should be located in the end of the file.
	MetadataPointer [8]byte
}

// Header identifying a DSD chunk within a DSD stream file.
const dsdChunkHeader = "DSD "

// Size in bytes of a DSD chunk within a DSD stream file.
const dsdChunkSize = 28

// readDSDChunk reads the DSD chunk and stores the result in d.
func (d *decoder) readDSDChunk() error {
	// Read the entire chunk in one go
	err := binary.Read(d.reader, binary.LittleEndian, &d.dsd)
	if err != nil {
		return err
	}

	// Chunk header
	header := string(d.dsd.Header[:])
	switch header {
	case dsdChunkHeader:
		// This is the expected chunk header
	case fmtChunkHeader:
		return fmt.Errorf("dsd: expected DSD chunk but found fmt chunk")
	case dataChunkHeader:
		return fmt.Errorf("dsd: expected DSD chunk but found data chunk")
	default:
		return fmt.Errorf("dsd: bad chunk header: %q\ndsd chunk: % x", header, d.dsd)
	}

	// Size of this chunk
	size := binary.LittleEndian.Uint64(d.dsd.Size[:])
	if size != dsdChunkSize {
		return fmt.Errorf("dsd: bad chunk size: %v bytes\ndsd chunk: % x", size, d.dsd)
	}

	// Total file size
	totalFileSize := binary.LittleEndian.Uint64(d.dsd.TotalFileSize[:])
	if totalFileSize < (dsdChunkSize + fmtChunkSize + dataChunkSize) {
		return fmt.Errorf("dsd: bad total file size: %v bytes\ndsd chunk: % x", totalFileSize, d.dsd)
	}

	// Pointer to Metadata chunk
	metadataPointer := binary.LittleEndian.Uint64(d.dsd.MetadataPointer[:])
	if metadataPointer != 0 {
		if metadataPointer >= totalFileSize || metadataPointer <= (dsdChunkSize+fmtChunkSize+dataChunkSize) {
			return fmt.Errorf("dsd: bad pointer to metadata chunk: %v bytes\ndsd chunk: % x", metadataPointer, d.dsd)
		} else {
			// Prepare the audio.Audio in d to hold the metadata
			d.audio.Metadata = make([]byte, totalFileSize-metadataPointer)
		}
	}

	// Log the fields of the chunk (only active if a log output has been set)
	d.logger.Print("\nDSD Chunk\n=========\n")
	d.logger.Printf("Chunk header:              %q\n", header)
	d.logger.Printf("Size of this chunk:        %v bytes\n", size)
	d.logger.Printf("Total file size:           %v bytes\n", totalFileSize)
	d.logger.Printf("Pointer to Metadata chunk: %v\n", metadataPointer)

	return nil
}

// writeDSDChunk writes the DSD chunk.
func (e *encoder) writeDSDChunk() error {
	// Chunk header
	header := dsdChunkHeader
	copy(e.dsd.Header[:], header)

	// Size of this chunk
	size := uint64(dsdChunkSize)
	binary.LittleEndian.PutUint64(e.dsd.Size[:], size)

	// Total file size
	totalFileSize := uint64(dsdChunkSize + fmtChunkSize + dataChunkSize +
		len(e.audio.EncodedSamples) + len(e.audio.Metadata))
	binary.LittleEndian.PutUint64(e.dsd.TotalFileSize[:], totalFileSize)

	// Pointer to Metadata chunk
	metadataPointer := uint64(0)
	if len(e.audio.Metadata) > 0 {
		metadataPointer = totalFileSize - uint64(len(e.audio.Metadata))
	}
	binary.LittleEndian.PutUint64(e.dsd.MetadataPointer[:], metadataPointer)

	// Log the fields of the chunk (only active if a log output has been set)
	e.logger.Print("\nDSD Chunk\n=========\n")
	e.logger.Printf("Chunk header:              %q\n", header)
	e.logger.Printf("Size of this chunk:        %v\n", size)
	e.logger.Printf("Total file size:           %v\n", totalFileSize)
	e.logger.Printf("Pointer to Metadata chunk: %v\n", metadataPointer)

	// Write the entire chunk in one go
	err := binary.Write(e.writer, binary.LittleEndian, &e.dsd)
	if err != nil {
		return err
	}

	return nil
}
