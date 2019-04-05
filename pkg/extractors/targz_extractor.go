package extractors

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/3p/archiver"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"io"
)

// knows how to extract from a tar.gz or a tar using archiver.Archiver interface
type tarGzExtractor struct {
	// the stream where the tar.gz bytes are read from
	readStream        io.Reader
	archiverInterface archiver.Archiver
}

func NewTarGzExtractor(reader io.Reader) Extractor {
	return newTarGzExtractor(reader, archiver.TarGz)
}

func NewTarExtractor(reader io.Reader) Extractor {
	return newTarGzExtractor(reader, archiver.Tar)
}

func newTarGzExtractor(reader io.Reader, archiverInterface archiver.Archiver) Extractor {
	return &tarGzExtractor{
		readStream:        reader,
		archiverInterface: archiverInterface,
	}
}

func ExtractorFromFileName(reader io.Reader, fileName string) Extractor {
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
	// crete the extract location if it doesn't exist
	extractLocationErr := fs.MkdirAll(extractLocation, DefaultFileMode)
	if extractLocationErr != nil {
		return extractLocationErr
	}

	// Now, extract the bytes
	extractErr := archiverInterface.Read(e.readStream, extractLocation)
	if extractErr != nil {
		return extractErr
	}

	return nil
}
