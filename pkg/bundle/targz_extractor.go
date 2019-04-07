package bundle

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:generate mockgen -destination=mock_file_system.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system FileSystem
//go:generate mockgen -destination=mock_file.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system File
//go:generate mockgen -destination=mock_file_info.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system FileInfo


import (
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/3p/archiver"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"io"
)

// knows how to Extract from a tar.gz or a tar using archiver.Archiver interface
type tarGzExtractor struct {
	// the stream where the tar.gz bytes are read from
	readStream        io.Reader
	archiverInterface archiver.Archiver
}

func newTarExtractor(reader io.Reader) *tarGzExtractor {
	return newExtractor(reader, archiver.Tar)
}

func newExtractor(reader io.Reader, archiverInterface archiver.Archiver) *tarGzExtractor {
	return &tarGzExtractor{
		readStream:        reader,
		archiverInterface: archiverInterface,
	}
}

func extractorFromFileName(reader io.Reader, fileName string) *tarGzExtractor {
	archiverInterface := archiver.MatchingFormat(fileName)

	if archiverInterface == nil {
		return nil
	}

	return &tarGzExtractor{
		readStream:        reader,
		archiverInterface: archiverInterface,
	}
}

func (e *tarGzExtractor) Extract(extractLocation string, fs file_system.FileSystem) error {
	return e.ExtractWithArchiver(extractLocation, fs, e.archiverInterface)
}

func (e *tarGzExtractor) ExtractWithArchiver(extractLocation string, fs file_system.FileSystem, archiverInterface archiver.Archiver) error {
	// crete the Extract location if it doesn't exist
	extractLocationErr := fs.MkdirAll(extractLocation, defaultFileMode)
	if extractLocationErr != nil {
		return extractLocationErr
	}

	// Now, Extract the bytes
	extractErr := archiverInterface.Read(e.readStream, extractLocation)
	if extractErr != nil {
		return extractErr
	}

	return nil
}
