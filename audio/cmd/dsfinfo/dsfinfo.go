// Copyright 2015 Simon Moore (simon@snmoore.net). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// dsfinfo reads a DSF (DSD Stream File) and prints information about its
// contents.
//
// Usage:
//		dsfinfo file
package main

import (
	"flag"
	"github.com/snmoore/go/audio/dsf"
	"io/ioutil"
	"os"
)

func main() {
	// The input file should be specified on the command line
	flag.Parse()
	filepath := flag.Arg(0)

	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}

	// Upon exit, close the file
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	// Decode the DSD stream file with logging to stdout
	audio, err := dsf.Decode(file, os.Stdout)
	if err != nil {
		panic(err)
	}

	// Encode the DSD stream file
	err = dsf.Encode(audio, ioutil.Discard, os.Stdout)
	if err != nil {
		panic(err)
	}
}
