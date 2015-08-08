// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package audio implements a basic audio library with support for the following
// audio file formats:
// 	DSF - DSD Stream File
package audio

// Encoding defines the set of possible audio encodings.
type Encoding int

const (
	// Direct Stream Digital (DSD) i.e. uncompressed DSD audio.
	DSD Encoding = iota

	// Direct Stream Transfer (DST) i.e. compressed DSD audio.
	DST
)

// Channel defines the set of possible audio channels.
type Channel int

const (
	FrontLeft Channel = iota
	FrontRight
	Center
	LowFrequency
	BackLeft
	BackRight
)

// Audio is a set of audio samples of a particular encoding.
type Audio struct {
	// The audio encoding e.g. DSD or DST.
	Encoding Encoding

	// The number of channels e.g. 2 for stereo.
	NumChannels uint

	// The channel order e.g. front left, front right.
	ChannelOrder []Channel

	// The sampling frequency in Hertz.
	SamplingFrequency uint

	// The number of bits per sample.
	BitsPerSample uint

	// Block size per channel in bytes.
	BlockSize uint

	// The encoded audio samples.
	EncodedSamples []byte

	// Metadata e.g. an ID3v2 tag.
	Metadata []byte
}

// String returns the lowercase name of a Channel.
func (c Channel) String() string {
	switch c {
	case FrontLeft:
		return "front left"
	case FrontRight:
		return "front right"
	case Center:
		return "center"
	case LowFrequency:
		return "low frequency"
	case BackLeft:
		return "back left"
	case BackRight:
		return "back right"
	}
	return "unknown"
}
