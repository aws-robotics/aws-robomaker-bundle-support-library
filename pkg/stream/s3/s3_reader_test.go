// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	testBucket      string = "testBucket"
	testKey         string = "testString"
	testEtag        string = "testEtag"
	testBodyContent string = "hello world"
)

type errReader struct {
}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

func setupS3MockExpects(mockS3Client *MockS3API) {
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(testBodyContent))),
		ETag:          aws.String(testEtag),
	}, nil).Times(1)

	mockS3Client.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.GetObjectOutput{
		Body: ioutil.NopCloser(strings.NewReader(testBodyContent[:2])),
	}, nil).Times(1)
}

func TestS3Reader_Read_ReadFails_Retries(t *testing.T) {
	const message = "Yo"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(testBodyContent))),
		ETag:          aws.String(testEtag),
	}, nil).Times(1)

	mockS3Client.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&s3.GetObjectOutput{
			Body: ioutil.NopCloser(&errReader{}),
		}, nil).Times(3)

	mockS3Client.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&s3.GetObjectOutput{
			Body: ioutil.NopCloser(strings.NewReader(message)),
		}, nil).Times(1)

	config := newS3ReaderConfig()
	config.RetryWait = 1 * time.Nanosecond
	config.NumRetries = 3
	s3Reader, _ := newS3ReaderWithConfig(mockS3Client, testBucket, testKey, config)

	content := make([]byte, 2)
	_, err := s3Reader.Read(content)
	assert.True(t, err == nil)
	assert.Equal(t, message, string(content))
}

func TestS3Reader_Read_ReadFails_ExhaustsAllRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(testBodyContent))),
		ETag:          aws.String(testEtag),
	}, nil).Times(1)

	mockS3Client.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&s3.GetObjectOutput{
			Body: ioutil.NopCloser(&errReader{}),
		}, nil).Times(4)

	config := newS3ReaderConfig()
	config.RetryWait = 1 * time.Nanosecond
	config.NumRetries = 3
	s3Reader, _ := newS3ReaderWithConfig(mockS3Client, testBucket, testKey, config)

	_, err := s3Reader.Read(make([]byte, 1))
	assert.True(t, err != nil)
}

func TestS3Reader_Read_CallFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(testBodyContent))),
		ETag:          aws.String(testEtag),
	}, nil).Times(1)

	mockS3Client.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&s3.GetObjectOutput{}, fmt.Errorf("Call failed")).Times(1)

	s3Reader, _ := newS3ReaderBucketAndKey(mockS3Client, testBucket, testKey)

	_, err := s3Reader.Read(make([]byte, 1))
	assert.True(t, err != nil)
}

func TestS3Reader_Read_ShouldBuffer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	setupS3MockExpects(mockS3Client)

	mockS3Client.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.GetObjectOutput{
		Body: ioutil.NopCloser(strings.NewReader(testBodyContent[2:4])),
	}, nil).Times(1)

	config := newS3ReaderConfig()
	config.BufferSize = 2
	s3Reader, _ := newS3ReaderWithConfig(mockS3Client, testBucket, testKey, config)

	//First call will read from s3 since the buffer is empty.
	//Second call should read from internal buffer
	//Third call should read from s3 again
	s3Reader.Read(make([]byte, 1))
	s3Reader.Read(make([]byte, 1))
	s3Reader.Read(make([]byte, 1))
}

func TestS3Reader_Seek_ResetsBuffer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	setupS3MockExpects(mockS3Client)
	s3Reader, _ := newS3ReaderBucketAndKey(mockS3Client, testBucket, testKey)

	//Read to fill the buffer
	s3Reader.Read(make([]byte, 1))
	assert.True(t, s3Reader.bufferEnd != 0)

	s3Reader.Seek(1, io.SeekCurrent)
	assert.True(t, s3Reader.bufferEnd == 0)
}

func TestS3Reader_SeekNoop_DoesNotResetBuffer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	setupS3MockExpects(mockS3Client)
	s3Reader, _ := newS3ReaderBucketAndKey(mockS3Client, testBucket, testKey)

	//Read to fill the buffer
	s3Reader.Read(make([]byte, 1))
	assert.True(t, s3Reader.bufferEnd != 0)

	s3Reader.Seek(0, io.SeekCurrent)
	assert.True(t, s3Reader.bufferEnd != 0)
}

func TestS3Reader_Seek_SeeksToCorrectPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(testBodyContent))),
		ETag:          aws.String(testEtag),
	}, nil).Times(1)
	s3Reader, _ := newS3ReaderBucketAndKey(mockS3Client, testBucket, testKey)

	s3Reader.Seek(5, io.SeekCurrent)
	assert.True(t, s3Reader.offset == 5)

	s3Reader.Seek(1, io.SeekCurrent)
	assert.True(t, s3Reader.offset == 6)

	s3Reader.Seek(0, io.SeekStart)
	assert.True(t, s3Reader.offset == 0)

	s3Reader.Seek(0, io.SeekEnd)
	assert.True(t, s3Reader.offset == s3Reader.ContentLength)

	s3Reader.Seek(100000, io.SeekCurrent)
	assert.True(t, s3Reader.offset == s3Reader.ContentLength)
}
