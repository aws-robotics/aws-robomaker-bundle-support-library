// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/extractors"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"io"
)

const (
	VersionFileName = "version"
)

// BundleArchive's responsibility is to be able to extract bundles of all versions, v1, v2, etc.
type BundleArchive interface {

	// What version is this bundle?
	Version() string

	// Extract everything into the cache
	Extract(bundleCache store.BundleStore) (bundle.Bundle, error)
}

type bundleArchive struct {
	version         string
	inputStream     io.ReadSeeker
	bundleProcessor BundleProcessor
}

func NewBundleArchive(inputStream io.ReadSeeker) (BundleArchive, error) {
	// read version to determine bundle version
	tarReader := extractors.TarReaderFromStream(inputStream)
	version, versionErr := ReadVersionFromBundle(tarReader)

	if versionErr != nil {
		return nil, fmt.Errorf("unable to read version from bundle: %v", versionErr)
	}

	// get the appropriate bundle processor for the version
	bundleProcessor := BundleProcessorForVersion(version)
	if bundleProcessor == nil {
		return nil, fmt.Errorf("unsuppported bundle processor version: %s", version)
	}

	// reset seek position to start of the stream and init with the stream
	inputStream.Seek(0, io.SeekStart)
	return &bundleArchive{
		version:         version,
		inputStream:     inputStream,
		bundleProcessor: bundleProcessor,
	}, nil
}

func (b *bundleArchive) Version() string {
	return b.version
}

func (b *bundleArchive) Extract(bundleStore store.BundleStore) (bundle.Bundle, error) {
	return b.bundleProcessor.Extract(b.inputStream, bundleStore)
}

func ReadVersionFromBundle(tarReader *tar.Reader) (string, error) {
	header, headerErr := tarReader.Next()
	if headerErr != nil {
		fmt.Printf("Error parsing tar: %v", headerErr)
		return "", headerErr
	}

	if header.Name != VersionFileName {
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
