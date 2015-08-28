// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package dsf

import (
	"encoding/binary"
	"fmt"
	"github.com/snmoore/go/audio"
	"reflect"
)

// FmtChunk is the file structure of the fmt chunk within a DSD stream file.
// See "DSF File Format Specification", v1.01, Sony Corporation. All data is
// little-endian. This is exported to allow reading with binary.Read.
type FmtChunk struct {
	// fmt chunk header.
	// 'f' , 'm' , 't' , ' ' (includes 1 space).
	Header [4]byte

	// Size of this chunk.
	// Usually 52 bytes.
	Size [8]byte

	// Format version.
	// 1.
	Version [4]byte

	// Format id.
	// 0: DSD raw.
	Identifier [4]byte

	// Channel type.
	// 1: mono
	// 2: stereo
	// 3: 3 channels
	// 4: quad
	// 5: 4 channels
	// 6: 5 channels
	// 7: 5.1 channels
	ChannelType [4]byte

	// Channel num.
	// 1: mono
	// 2: stereo
	// ...
	// 6: 6 channels
	ChannelNum [4]byte

	// Sampling frequency in Hertz.
	// 2822400, 5644800, 11289600 or 22579200.
	SamplingFrequency [4]byte

	// Bits per sample.
	// 1 or 8.
	BitsPerSample [4]byte

	// Sample count.
	// This is for 1 channel e.g. for n seconds of data:
	// SampleCount = SamplingFrequency * n.
	SampleCount [8]byte

	// Block size per channel.
	// 4096, unused samples should be filled with zero.
	BlockSize [4]byte

	// Reserved.
	// Filled with zero.
	Reserved [4]byte
}

// Header identifying a fmt chunk within a DSD stream file.
const fmtChunkHeader = "fmt "

// Size in bytes of a fmt chunk within a DSD stream file.
const fmtChunkSize = 52

// Value of the Version field.
const fmtVersion = 1

// Value of the Identifier field.
const fmtIdentifier = 0

// Values of the ChannelType field and their meaning.
var fmtChannelType = map[uint32]string{
	1: "mono",
	2: "stereo",
	3: "3 channels",
	4: "quad",
	5: "4 channels",
	6: "5 channels",
	7: "5.1 channels",
}

// Channel order corresponding to the ChannelType field.
// The mapping for mono is undefined in the specification, but using center
// seems reasonable and allows an easy way to check for mismatch between the
// ChannelType and ChannelNum fields.
var fmtChannelOrder = map[uint32][]audio.Channel{
	1: {audio.Center},
	2: {audio.FrontLeft, audio.FrontRight},
	3: {audio.FrontLeft, audio.FrontRight, audio.Center},
	4: {audio.FrontLeft, audio.FrontRight, audio.BackLeft, audio.BackRight},
	5: {audio.FrontLeft, audio.FrontRight, audio.Center, audio.LowFrequency},
	6: {audio.FrontLeft, audio.FrontRight, audio.Center, audio.BackLeft, audio.BackRight},
	7: {audio.FrontLeft, audio.FrontRight, audio.Center, audio.LowFrequency, audio.BackLeft, audio.BackRight},
}

// Values of the ChannelNum field and their meaning.
var fmtChannelNum = map[uint32]string{
	1: "mono",
	2: "stereo",
	3: "3 channels",
	4: "4 channels",
	5: "5 channels",
	6: "6 channels",
}

// Values of the SamplingFrequency field and their meaning.
// Only 2822400 and 5644800 are defined by the specification, but the other
// rates are in active use. The strings are not defined within the specification
// but are in active use.
var fmtSamplingFrequency = map[uint32]string{
	2822400:  "DSD64",
	5644800:  "DSD128",
	11289600: "DSD256",
	22579200: "DSD512",
}

// Values of the BitsPerSample field.
var fmtBitsPerSample = map[uint32]struct{}{
	1: {},
	8: {},
}

// Value of the BlockSize field.
const fmtBlockSize = 4096

// Value of the Reserved field.
const fmtReserved = 0

// readFmtChunk reads the fmt chunk and stores the result in d.
func (d *decoder) readFmtChunk() error {
	// Read the entire chunk in one go
	err := binary.Read(d.reader, binary.LittleEndian, &d.fmt)
	if err != nil {
		return err
	}

	// Chunk header
	header := string(d.fmt.Header[:])
	switch header {
	case fmtChunkHeader:
		// This is the expected chunk header
	case dsdChunkHeader:
		return fmt.Errorf("fmt: expected fmt chunk but found DSD chunk")
	case dataChunkHeader:
		return fmt.Errorf("fmt: expected fmt chunk but found data chunk")
	default:
		return fmt.Errorf("fmt: bad chunk header: %q\nfmt chunk: % x", header, d.fmt)
	}

	// Size of this chunk
	size := binary.LittleEndian.Uint64(d.fmt.Size[:])
	if size != fmtChunkSize {
		return fmt.Errorf("fmt: bad chunk size: %v\nfmt chunk: % x", size, d.fmt)
	}

	// Format version
	formatVersion := binary.LittleEndian.Uint32(d.fmt.Version[:])
	if formatVersion != fmtVersion {
		return fmt.Errorf("fmt: bad format version: %v\nfmt chunk: % x", formatVersion, d.fmt)
	}

	// Format id
	formatId := binary.LittleEndian.Uint32(d.fmt.Identifier[:])
	if formatId != fmtIdentifier {
		return fmt.Errorf("fmt: bad format id: %v\nfmt chunk: % x", formatId, d.fmt)
	}

	// Channel Type
	channelType := binary.LittleEndian.Uint32(d.fmt.ChannelType[:])
	channelTypeString, ok := fmtChannelType[channelType]
	if !ok {
		return fmt.Errorf("fmt: bad channel type: %v\nfmt chunk: % x", channelType, d.fmt)
	}

	// Channel order corresponding to the ChannelType field
	order, _ := fmtChannelOrder[channelType]

	// Channel num
	channelNum := binary.LittleEndian.Uint32(d.fmt.ChannelNum[:])
	_, ok = fmtChannelNum[channelNum]
	if !ok {
		return fmt.Errorf("fmt: bad channel num: %v\nfmt chunk: % x", channelNum, d.fmt)
	}
	if channelNum != uint32(len(order)) {
		return fmt.Errorf("fmt: mismatch between channel type %v and channel num %v:\nfmt chunk: % x", channelType, channelNum, d.fmt)
	}

	// Sampling frequency
	samplingFrequency := binary.LittleEndian.Uint32(d.fmt.SamplingFrequency[:])
	samplingFrequencyString, ok := fmtSamplingFrequency[samplingFrequency]
	if !ok {
		return fmt.Errorf("fmt: bad sampling frequency: %v\nfmt chunk: % x", samplingFrequency, d.fmt)
	}

	// Bits per sample
	bitsPerSample := binary.LittleEndian.Uint32(d.fmt.BitsPerSample[:])
	_, ok = fmtBitsPerSample[bitsPerSample]
	if !ok {
		return fmt.Errorf("fmt: bad bits per sample: %v\nfmt chunk: % x", bitsPerSample, d.fmt)
	}

	// Sample count
	sampleCount := binary.LittleEndian.Uint64(d.fmt.SampleCount[:])

	// Block size per channel
	blockSize := binary.LittleEndian.Uint32(d.fmt.BlockSize[:])
	if blockSize != fmtBlockSize {
		return fmt.Errorf("fmt: bad block size: %v\nfmt chunk: % x", blockSize, d.fmt)
	}

	// Reserved
	reserved := binary.LittleEndian.Uint32(d.fmt.Reserved[:])
	if reserved != fmtReserved {
		return fmt.Errorf("fmt: bad reserved bytes: %#x\nfmt chunk: % x", reserved, d.fmt)
	}

	// Log the fields of the chunk (only active if a log output has been set)
	d.logger.Print("\nFmt Chunk\n=========\n")
	d.logger.Printf("Chunk header:              %q\n", header)
	d.logger.Printf("Size of this chunk:        %v bytes\n", size)
	d.logger.Printf("Format version:            %v\n", formatVersion)
	d.logger.Printf("Format id:                 %v\n", formatId)
	d.logger.Printf("Channel type:              %v (%s)\n", channelType, channelTypeString)
	d.logger.Printf("Channel num:               %v\n", channelNum)
	if len(order) > 1 {
		var s string
		for i, channel := range order {
			if i < len(order)-1 {
				s += channel.String() + ", "
			} else {
				s += channel.String()
			}
		}
		d.logger.Printf("Channel order:             %v\n", s)
	}
	d.logger.Printf("Sampling frequency:        %vHz (%s)\n", samplingFrequency, samplingFrequencyString)
	d.logger.Printf("Bits per sample:           %v\n", bitsPerSample)
	d.logger.Printf("Sample count:              %v\n", sampleCount)
	d.logger.Printf("Block size per channel:    %v bytes\n", blockSize)

	// Store the information that is useful
	d.audio.Encoding = audio.DSD
	d.audio.NumChannels = uint(channelNum)
	d.audio.ChannelOrder = order
	d.audio.SamplingFrequency = uint(samplingFrequency)
	d.audio.BitsPerSample = uint(bitsPerSample)
	d.audio.BlockSize = uint(blockSize)

	// Prepare the audio.Audio in d to hold the encoded samples
	length := sampleCount / (8 / uint64(bitsPerSample))        // number of bytes per channel
	length += uint64(blockSize) - (length % uint64(blockSize)) // round up to the block size
	length *= uint64(channelNum)                               // number of channels
	d.audio.EncodedSamples = make([]byte, length)

	return nil
}

// writeFmtChunk writes the fmt chunk.
func (e *encoder) writeFmtChunk() error {
	// Chunk header
	header := fmtChunkHeader
	copy(e.fmt.Header[:], header)

	// Size of this chunk
	size := uint64(fmtChunkSize)
	binary.LittleEndian.PutUint64(e.fmt.Size[:], size)

	// Format version
	formatVersion := uint32(fmtVersion)
	binary.LittleEndian.PutUint32(e.fmt.Version[:], formatVersion)

	// Format id
	formatId := uint32(fmtIdentifier)
	binary.LittleEndian.PutUint32(e.fmt.Identifier[:], formatId)

	// Channel type
	var channelType uint32
	for key, order := range fmtChannelOrder {
		if reflect.DeepEqual(e.audio.ChannelOrder, order) {
			channelType = key
		}
	}
	if channelType == 0 {
		var s string
		for i, channel := range e.audio.ChannelOrder {
			if i < len(e.audio.ChannelOrder)-1 {
				s += channel.String() + ", "
			} else {
				s += channel.String()
			}
		}
		return fmt.Errorf("fmt: unsupported channel ordering: %v", s)
	}
	channelTypeString, _ := fmtChannelType[channelType]
	binary.LittleEndian.PutUint32(e.fmt.ChannelType[:], channelType)

	// Channel num
	channelNum := uint32(e.audio.NumChannels)
	if channelNum > 1 && (channelNum != uint32(len(e.audio.ChannelOrder))) {
		return fmt.Errorf("fmt: mismatch between num channels and channel order: %v, %v", channelNum, e.audio.ChannelOrder)
	}
	binary.LittleEndian.PutUint32(e.fmt.ChannelNum[:], channelNum)

	// SamplingFrequency
	samplingFrequency := uint32(e.audio.SamplingFrequency)
	samplingFrequencyString, ok := fmtSamplingFrequency[samplingFrequency]
	if !ok {
		return fmt.Errorf("fmt: unsupported sampling frequency: %v", samplingFrequency)
	}
	binary.LittleEndian.PutUint32(e.fmt.SamplingFrequency[:], samplingFrequency)

	// Bits per sample
	bitsPerSample := uint32(e.audio.BitsPerSample)
	_, ok = fmtBitsPerSample[bitsPerSample]
	if !ok {
		return fmt.Errorf("fmt: unsupported bits per sample: %v", bitsPerSample)
	}
	binary.LittleEndian.PutUint32(e.fmt.BitsPerSample[:], bitsPerSample)

	// SampleCount

	// Log the fields of the chunk (only active if a log output has been set)
	e.logger.Print("\nFmt Chunk\n=========\n")
	e.logger.Printf("Chunk header:              %q\n", header)
	e.logger.Printf("Size of this chunk:        %v\n", size)
	e.logger.Printf("Format version:            %v\n", formatVersion)
	e.logger.Printf("Format id:                 %v\n", formatId)
	e.logger.Printf("Channel type:              %v (%s)\n", channelType, channelTypeString)
	e.logger.Printf("Channel num:               %v\n", channelNum)
	if len(e.audio.ChannelOrder) > 1 {
		var s string
		for i, channel := range e.audio.ChannelOrder {
			if i < len(e.audio.ChannelOrder)-1 {
				s += channel.String() + ", "
			} else {
				s += channel.String()
			}
		}
		e.logger.Printf("Channel order:             %v\n", s)
	}
	e.logger.Printf("Sampling frequency:        %vHz (%s)\n", samplingFrequency, samplingFrequencyString)
	e.logger.Printf("Bits per sample:           %v\n", bitsPerSample)

	// Write the entire chunk in one go
	err := binary.Write(e.writer, binary.LittleEndian, &e.fmt)
	if err != nil {
		return err
	}

	return nil
}
