// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package stream provides support to convert a URL into an io.ReadSeeker.
//
// Streamers should be registered via RegisterStreamer() in order
// to use them with UrlToStream()
package stream

import (
	"fmt"
	"io"
)

var streamers []Streamer

// Streamer is the interface implemented by an object that can create an io.ReadSeeker from a remote or local URL
type Streamer interface {
	// Returns true if this streamer can create a stream from the url
	CanStream(url string) bool
	// Opens a stream to the contents of url
	// Returns:
	// io.ReadSeeker - data at url
	// Length of the stream in bytes
	// checksum of the file pointed to by path
	// error if any
	CreateStream(url string) (io.ReadSeeker, int64, string, error)
}

// UrlToStream converts a URL into an io.ReadSeeker
// returns:
// the io.ReadSeeker
// the length of the stream in bytes
// checksum of the file pointed to by path
// error if any
func UrlToStream(url string) (io.ReadSeeker, int64, string, error) {
	for i := 0; i < len(streamers); i++ {
		if streamers[i].CanStream(url) {
			return streamers[i].CreateStream(url)
		}
	}
	return nil, 0, "", fmt.Errorf("no supported Streamer was found for %s", url)
}

// Adds the Streamer to the list of handlers used by UrlToStream method
func RegisterStreamer(s Streamer) {
	streamers = append(streamers, s)
}
