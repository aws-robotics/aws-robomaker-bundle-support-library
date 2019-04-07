// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"archive/tar"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"io"
)

const (
	metadataFileName = "metadata.tar"
	bundleFileName   = "bundle.tar"
)

var expectedFiles = [...]string{metadataFileName, bundleFileName}

// Knows how to Extract all files from a v1 bundle
type v1Extractor struct {

	// the stream where the bundle's bytes are read from
	readStream io.ReadSeeker
}

func newBundleV1Extractor(reader io.ReadSeeker) *v1Extractor {
	return &v1Extractor{
		readStream: reader,
	}
}

func (e *v1Extractor) Extract(extractLocation string, fs file_system.FileSystem) error {
	return e.extractWithTarReader(tarReaderFromStream(e.readStream), extractLocation, fs)
}

func (e *v1Extractor) extractWithTarReader(tarReader *tar.Reader, extractLocation string, fs file_system.FileSystem) error {
	// crete the Extract location if it doesn't exist
	extractLocationErr := fs.MkdirAll(extractLocation, defaultFileMode)
	if extractLocationErr != nil {
		return extractLocationErr
	}

	// iterate headers and process each file in the tar
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			// we there are no more headers, we finish
			break
		} else if err != nil {
			return err
		}

		// we only Extract when they are expected files
		if isExpectedFile(header.Name) {
			extractErr := newTarExtractor(tarReader).Extract(extractLocation, fs)
			if extractErr != nil {
				return extractErr
			}
		}
	}
	return nil
}

func isExpectedFile(fileName string) bool {
	for _, tarFile := range expectedFiles {
		if fileName == tarFile {
			return true
		}
	}
	return false
}
