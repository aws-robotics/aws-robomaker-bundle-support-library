// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stream

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.amzn.com/robomaker/bundle_support/file_system"
	"testing"
)

type ofHeadObjectInput struct {
	Bucket string
	Key    string
}

func OfHeadObjectInput(bucket string, key string) gomock.Matcher {
	return &ofHeadObjectInput{Bucket: bucket, Key: key}
}

func (o *ofHeadObjectInput) Matches(x interface{}) bool {
	that, ok := x.(*s3.HeadObjectInput)
	if !ok {
		return false
	}
	return *that.Bucket == o.Bucket && *that.Key == o.Key
}

func (o *ofHeadObjectInput) String() string {
	return "Bucket: " + o.Bucket + " Key: " + o.Key
}

func TestPathToStream_WithLocalFile_ShouldReturnStreamAndMd5(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedMd5 := "12345"
	filePath := "/test/stream"

	mockS3Client := NewMockS3API(ctrl)
	mockFile := file_system.NewMockFile(ctrl)
	mockFileSystem := file_system.NewMockFileSystem(ctrl)
	mockFileInfo := file_system.NewMockFileInfo(ctrl)
	mockFileSystem.EXPECT().Open(filePath).Return(mockFile, nil)
	mockFile.EXPECT().Stat().Return(mockFileInfo, nil)
	mockFileInfo.EXPECT().Size().Return(int64(1))

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		return expectedMd5, nil
	}

	stream, md5, contentLength, err := pathToStream(filePath, mockS3Client, mockFileSystem, computeMd5)

	assert.Equal(t, mockFile, stream)
	assert.Equal(t, expectedMd5, md5)
	assert.Equal(t, int64(1), contentLength)
	assert.Nil(t, err)
}

func TestPathToStream_WithInvalidLocalFile_ShouldReturnError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "/test/stream"
	expectedErr := fmt.Errorf("error")

	mockS3Client := NewMockS3API(ctrl)
	mockFileSystem := file_system.NewMockFileSystem(ctrl)
	mockFileSystem.EXPECT().Open(filePath).Return(nil, expectedErr)

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		return "", nil
	}

	stream, md5, _, err := pathToStream(filePath, mockS3Client, mockFileSystem, computeMd5)

	assert.Nil(t, stream)
	assert.Equal(t, "", md5)
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestPathToStream_WithS3Url_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "s3://test/stream"
	err := fmt.Errorf("test error")

	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(nil, err)
	mockFileSystem := file_system.NewMockFileSystem(ctrl)

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		assert.Fail(t, "computeMd5 should not be called.")
		return "", nil
	}

	stream, md5, _, err := pathToStream(filePath, mockS3Client, mockFileSystem, computeMd5)

	assert.Nil(t, stream)
	assert.Equal(t, "", md5)
	assert.NotNil(t, err)
}

func TestPathToStream_WithValidS3URL_ShouldReturnValidStream(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// s3etag is wrapped in quotes
	s3Etag := "\"abcdefg\""

	// we expect our final etag to have quotes stripped
	expectedEtag := "\"abcdefg\""

	var expectedContentLength int64 = 12345

	expectedHeadObjectOutput := &s3.HeadObjectOutput{
		ETag:          &s3Etag,
		ContentLength: &expectedContentLength,
	}

	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(OfHeadObjectInput("test", "stream")).Return(expectedHeadObjectOutput, nil)

	mockFileSystem := file_system.NewMockFileSystem(ctrl)

	filePath := "s3://test/stream"

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		assert.Fail(t, "computeMd5 should not be called.")
		return "", nil
	}

	stream, etag, contentLength, err := pathToStream(filePath, mockS3Client, mockFileSystem, computeMd5)

	assert.NotNil(t, stream)
	assert.Equal(t, expectedEtag, etag)
	assert.Equal(t, expectedContentLength, contentLength)
	assert.Nil(t, err)
}

func TestPathToStream_WithS3UrlWithoutClientAndWithoutRegion_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "s3://test/stream"
	err := fmt.Errorf("test error")

	mockFileSystem := file_system.NewMockFileSystem(ctrl)

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		assert.Fail(t, "computeMd5 should not be called.")
		return "", nil
	}

	stream, md5, _, err := pathToStream(filePath, nil, mockFileSystem, computeMd5)

	assert.Nil(t, stream)
	assert.Equal(t, "", md5)
	assert.NotNil(t, err)
}

func TestPathToStream_WithHttpUrl_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "http://test/stream"

	mockS3Client := NewMockS3API(ctrl)
	mockFileSystem := file_system.NewMockFileSystem(ctrl)

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		assert.Fail(t, "computeMd5 should not be called.")
		return "", nil
	}

	stream, md5, _, err := pathToStream(filePath, mockS3Client, mockFileSystem, computeMd5)

	assert.Nil(t, stream)
	assert.Equal(t, "", md5)
	assert.NotNil(t, err)
}

func TestPathToStream_WithHttpsUrl_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "https://test/stream"

	mockS3Client := NewMockS3API(ctrl)
	mockFileSystem := file_system.NewMockFileSystem(ctrl)

	computeMd5 := func(filePath string, fileSystem file_system.FileSystem) (string, error) {
		assert.Fail(t, "computeMd5 should not be called.")
		return "", nil
	}

	stream, md5, _, err := pathToStream(filePath, mockS3Client, mockFileSystem, computeMd5)

	assert.Nil(t, stream)
	assert.Equal(t, "", md5)
	assert.NotNil(t, err)
}
