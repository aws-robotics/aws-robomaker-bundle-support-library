// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	extractLocation              = "/extractLocation"
	expectedFileMode os.FileMode = 0755
)

func TestTarGzExtractor_Extract_WithNoErrors_ShouldExtract(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchiver := NewMockArchiver(ctrl)
	mockFileSystem := NewMockFs(ctrl)

	mockFileSystem.EXPECT().MkdirAll(extractLocation, expectedFileMode).Return(nil)
	mockArchiver.EXPECT().Read(nil, extractLocation).Return(nil)

	extractor := tarGzExtractor{}
	extractErr := extractor.ExtractWithArchiver(extractLocation, mockFileSystem, mockArchiver)

	assert.Nil(t, extractErr)
}

func TestTarGzExtractor_Extract_WithMkDirAllErrors_ShouldErrorAndNotExtract(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchiver := NewMockArchiver(ctrl)
	mockFileSystem := NewMockFs(ctrl)

	mkdirAllError := errors.New("MkDirAll Error")

	mockFileSystem.EXPECT().MkdirAll(extractLocation, expectedFileMode).Return(mkdirAllError)

	extractor := tarGzExtractor{}
	extractErr := extractor.ExtractWithArchiver(extractLocation, mockFileSystem, mockArchiver)

	assert.NotNil(t, extractErr)
	assert.Equal(t, mkdirAllError, extractErr)
}

func TestTarGzExtractor_Extract_WithExtractErrors_ShouldError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchiver := NewMockArchiver(ctrl)
	mockFileSystem := NewMockFs(ctrl)

	tarGzErr := errors.New("tarGzErr")

	mockFileSystem.EXPECT().MkdirAll(extractLocation, expectedFileMode).Return(nil)
	mockArchiver.EXPECT().Read(nil, extractLocation).Return(tarGzErr)

	extractor := tarGzExtractor{}
	extractErr := extractor.ExtractWithArchiver(extractLocation, mockFileSystem, mockArchiver)

	assert.NotNil(t, extractErr)
	assert.Equal(t, tarGzErr, extractErr)
}
