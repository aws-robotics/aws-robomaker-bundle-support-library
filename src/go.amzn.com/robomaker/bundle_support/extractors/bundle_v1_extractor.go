// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extractors

import (
	"archive/tar"
	"go.amzn.com/robomaker/bundle_support/file_system"
	"io"
)

const (
	metadataFileName = "metadata.tar"
	bundleFileName   = "bundle.tar"
)

var expectedFiles = [...]string{metadataFileName, bundleFileName}

// Knows how to extract all files from a v1 bundle
type BundleV1Extractor struct {

	// the stream where the bundle's bytes are read from
	readStream io.ReadSeeker
}

func NewBundleV1Extractor(reader io.ReadSeeker) Extractor {
	return &BundleV1Extractor{
		readStream: reader,
	}
}

func (e *BundleV1Extractor) Extract(extractLocation string, fs file_system.FileSystem) error {
	return e.ExtractWithTarReader(TarReaderFromStream(e.readStream), extractLocation, fs)
}

func (e *BundleV1Extractor) ExtractWithTarReader(tarReader *tar.Reader, extractLocation string, fs file_system.FileSystem) error {
	// crete the extract location if it doesn't exist
	extractLocationErr := fs.MkdirAll(extractLocation, DefaultFileMode)
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

		// we only extract when they are expected files
		if isExpectedFile(header.Name) {
			extractErr := NewTarExtractor(tarReader).Extract(extractLocation, fs)
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
