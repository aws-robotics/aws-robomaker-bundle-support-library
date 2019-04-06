// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extractors

//go:generate mockgen -destination=mock_extractor.go -package=extractors github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/extractors Extractor
//go:generate mockgen -destination=mock_archiver.go -package=extractors github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/3p/archiver Archiver

import (
	"archive/tar"
	"compress/gzip"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"io"
)

const (
	DefaultFileMode file_system.FileMode = 0755
)

// Extractor's responsibility is to extract all its contents into the target extract location
type Extractor interface {
	Extract(extractLocation string, fs file_system.FileSystem) error
}

func TarReaderFromStream(inputStream io.ReadSeeker) *tar.Reader {
	var tarReader *tar.Reader = nil

	// try as a gzReader
	gzReader, gzErr := gzip.NewReader(inputStream)
	if gzErr == nil {
		// this is a valid gz file
		// create the tar reader from the gzReader
		tarReader = tar.NewReader(gzReader)
	} else {
		// it wasn't a gz file, let's try to create the tar reader from the input stream
		inputStream.Seek(0, io.SeekStart)
		tarReader = tar.NewReader(inputStream)
	}
	return tarReader
}
