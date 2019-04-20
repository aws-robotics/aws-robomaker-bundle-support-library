// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

//go:generate mockgen -destination=mock_archiver.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/3p/archiver Archiver

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/fs"
	"io"
)

const (
	versionFileName             = "version"
	defaultFileMode fs.FileMode = 0755
)

// archive's responsibility is to be able to Extract bundles of all versions, v1, v2, etc.
type archive struct {
	version         string
	inputStream     io.ReadSeeker
	bundleProcessor bundleProcessor
}

func newBundleArchive(inputStream io.ReadSeeker) (*archive, error) {
	// read version to determine bundle version
	tarReader := tarReaderFromStream(inputStream)
	version, versionErr := readVersionFromBundle(tarReader)

	if versionErr != nil {
		return nil, fmt.Errorf("unable to read version from bundle: %v", versionErr)
	}

	// get the appropriate bundle processor for the version
	bundleProcessor := processorForVersion(version)
	if bundleProcessor == nil {
		return nil, fmt.Errorf("unsuppported bundle processor version: %s", version)
	}

	// reset seek position to start of the stream and init with the stream
	inputStream.Seek(0, io.SeekStart)
	return &archive{
		version:         version,
		inputStream:     inputStream,
		bundleProcessor: bundleProcessor,
	}, nil
}

// What version is this bundle?
func (b *archive) Version() string {
	return b.version
}

// Extract everything into the cache
func (b *archive) Extract(bundleStore Cache) (Bundle, error) {
	return b.bundleProcessor.extract(b.inputStream, bundleStore)
}

func readVersionFromBundle(tarReader *tar.Reader) (string, error) {
	header, headerErr := tarReader.Next()
	if headerErr != nil {
		fmt.Printf("bundleError parsing tar: %v", headerErr)
		return "", headerErr
	}

	if header.Name != versionFileName {
		err := errors.New("invalid bundle format, first file should be a version file")
		return "", err
	}

	versionData := make([]byte, header.Size)
	_, readVersionErr := tarReader.Read(versionData)
	// We need to read a second time to get the io.EOF message
	_, readVersionErr = tarReader.Read(nil)
	if readVersionErr != io.EOF {
		return "", fmt.Errorf("unable to read version: %v", readVersionErr)
	}

	return string(versionData), nil
}

func tarReaderFromStream(inputStream io.ReadSeeker) *tar.Reader {
	var tarReader *tar.Reader

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
