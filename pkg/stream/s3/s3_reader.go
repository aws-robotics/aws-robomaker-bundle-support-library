// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package s3

//go:generate mockgen -destination=mock_s3.go -package=s3 github.com/aws/aws-sdk-go/service/s3/s3iface S3API

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type s3ReaderConfig struct {
	NumRetries int
	RetryWait  time.Duration
	BufferSize int64
}

func newS3ReaderConfig() s3ReaderConfig {
	//Retry up to 20 seconds at 5 second intervals
	return s3ReaderConfig{
		NumRetries: 4,
		RetryWait:  time.Duration(5) * time.Second,
		BufferSize: 5 * 1024 * 1024, //5MB,
	}
}

//Implements io.ReadSeeker
type s3Reader struct {
	config        s3ReaderConfig
	buffer        []byte
	bufferStart   int64
	bufferEnd     int64
	bucket        string
	key           string
	s3            s3iface.S3API
	offset        int64
	ContentLength int64
	Etag          string
}

func newS3Reader(s3Api s3iface.S3API, bucket string, key string, contentLength int64, etag string, config s3ReaderConfig) *s3Reader {
	return &s3Reader{
		buffer:        make([]byte, config.BufferSize),
		bufferStart:   0,
		bufferEnd:     0,
		bucket:        bucket,
		key:           key,
		s3:            s3Api,
		offset:        0,
		ContentLength: contentLength,
		Etag:          etag,
		config:        config,
	}
}

type S3ReadError struct {
	err error
}

func (e *S3ReadError) Error() string {
	return e.err.Error()
}

func newS3ReaderBucketAndKey(s3Api s3iface.S3API, bucket string, key string) (*s3Reader, error) {
	return newS3ReaderWithConfig(s3Api, bucket, key, newS3ReaderConfig())
}

func newS3ReaderWithConfig(s3Api s3iface.S3API, bucket string, key string, config s3ReaderConfig) (*s3Reader, error) {
	resp, err := s3Api.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	return newS3Reader(s3Api, bucket, key, *resp.ContentLength, *resp.ETag, config), nil
}

/*
 * Read up to len(p) bytes by peforming a Ranged Get request.
 * This is useful for large objects and/or spotty connections,
 * where connection issues may occur when reading from the Body
 */
func (r *s3Reader) Read(p []byte) (n int, err error) {
	//Retry to handle dropped / spotty connections
	//AWS SDK retry strategy will only handle failed API calls, but not failed reads on the underlying stream
	//AWS SDK will also not retry on client errors, e.g. no network connection is present
	for i := r.config.NumRetries; i >= 0; i-- {
		n, err = r.read(p)

		//Retry on all request errors - worst case this adds 20 seconds to the deployment
		shouldRetry := false
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "RequestError" {
				shouldRetry = true
			}
		}

		if _, ok := err.(*S3ReadError); ok {
			shouldRetry = true
		}

		if shouldRetry {
			fmt.Printf("Error in s3Reader.Read: (%v). Retrying...\n", err)
			time.Sleep(r.config.RetryWait)
			continue
		}

		break
	}

	return
}

func (r *s3Reader) read(p []byte) (n int, err error) {
	if r.bufferEnd != 0 && r.bufferStart != r.bufferEnd {
		return r.copyInto(p)
	}

	r.resetBuffer()

	bufferBytes := int64(len(r.buffer))
	bytesLeft := r.ContentLength - r.offset
	bytesToRead := min(bufferBytes, bytesLeft)

	resp, getObjectErr := r.s3.GetObjectWithContext(
		aws.BackgroundContext(),
		&s3.GetObjectInput{
			Bucket:  aws.String(r.bucket),
			Key:     aws.String(r.key),
			IfMatch: aws.String(r.Etag),
			Range:   aws.String(fmt.Sprintf("bytes=%v-%v", r.offset, r.offset+bytesToRead-1)),
		},
		request.WithResponseReadTimeout(5*time.Second)) //Sets a timeout on the underlying Body.Read() calls. By default, there is no timeout on this read

	if getObjectErr != nil {
		return 0, getObjectErr
	}

	defer resp.Body.Close()
	for {
		bytesRead, readErr := resp.Body.Read(r.buffer[r.bufferEnd:])
		r.bufferEnd += int64(bytesRead)

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return n, &S3ReadError{readErr}
		}
	}

	return r.copyInto(p)
}

func (r *s3Reader) copyInto(p []byte) (n int, err error) {
	n = copy(p, r.buffer[r.bufferStart:r.bufferEnd])
	r.bufferStart += int64(n)

	r.offset += int64(n)
	if r.offset == r.ContentLength {
		return n, io.EOF
	}

	return n, nil
}

func (r *s3Reader) resetBuffer() {
	r.bufferStart = 0
	r.bufferEnd = 0
}

func (r *s3Reader) Seek(offset int64, whence int) (newOffset int64, err error) {
	oldPos := r.offset

	switch whence {
	default:
		return 0, fmt.Errorf("Seek: invalid whence %v", whence)
	case io.SeekStart:
		r.offset = offset
	case io.SeekCurrent:
		r.offset += offset
	case io.SeekEnd:
		r.offset = r.ContentLength
	}

	if r.offset > r.ContentLength {
		r.offset = r.ContentLength
	}

	//Reset buffer when seeking
	//Special case: Dont reset if the position hasn't changed
	if oldPos != r.offset {
		r.resetBuffer()
	}

	return r.offset, nil
}

func min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}
