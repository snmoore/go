# Go packages and commands

This contains work in progress - see [TODO.md](/TODO.md) for details.

## Package audio
[![GoDoc](https://godoc.org/github.com/snmoore/go/audio?status.svg)](https://godoc.org/github.com/snmoore/go/audio)

Implements a basic audio library with support for the following audio file formats:

* DSF - DSD Stream File

## Package audio/dsf
[![GoDoc](https://godoc.org/github.com/snmoore/go/audio/dsf?status.svg)](https://godoc.org/github.com/snmoore/go/audio/dsf)

Implements a DSF (DSD Stream File) audio decoder and encoder.

See "DSF File Format Specification", v1.01, Sony Corporation: http://dsd-guide.com/sites/default/files/white-papers/DSFFileFormatSpec_E.pdf

## Command dsfinfo
[![GoDoc](https://godoc.org/github.com/snmoore/go/audio/cmd/dsfinfo?status.svg)](https://godoc.org/github.com/snmoore/go/audio/cmd/dsfinfo)

Reads a DSF (DSD Stream File) and prints information about its contents.
 
    Usage:
        dsfinfo file
