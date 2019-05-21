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
	resp          *s3.GetObjectOutput
	bucket        string
	key           string
	s3            s3iface.S3API
	offset        int64
	ContentLength int64
	Etag          string
}

func newS3Reader(s3Api s3iface.S3API, bucket string, key string, contentLength int64, etag string, config s3ReaderConfig) *s3Reader {
	return &s3Reader{
		bucket:        bucket,
		key:           key,
		s3:            s3Api,
		offset:        0,
		ContentLength: contentLength,
		Etag:          etag,
		config:        config,
	}
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

		if _, ok := err.(*ReadError); ok {
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
	if r.resp == nil {
		getObjectErr := r.makeNewS3Request()
		if getObjectErr != nil {
			return 0, getObjectErr
		}
	}

	bytesRead, readErr := r.resp.Body.Read(p)

	if readErr != nil {
		// Error throwing from S3, close the current s3 socket
		defer r.closeS3Socket()
		if readErr == io.EOF {
			r.offset += int64(bytesRead)
			return bytesRead, io.EOF
		} else {
			return bytesRead, &ReadError{readErr}
		}
	} else {
		r.offset += int64(bytesRead)
		return bytesRead, nil

	}
}

func (r *s3Reader) makeNewS3Request() (err error) {
	resp, getObjectErr := r.s3.GetObjectWithContext(
		aws.BackgroundContext(),
		&s3.GetObjectInput{
			Bucket:  aws.String(r.bucket),
			Key:     aws.String(r.key),
			IfMatch: aws.String(r.Etag),
			// Always open a connection read from current position to end of file
			Range: aws.String(fmt.Sprintf("bytes=%v-%v", r.offset, r.ContentLength-1)),
		},
		// Sets a timeout on the underlying Body.Read() calls to avoid a S3 connection leaking
		// since this S3Reader keep connection open until read end of file. This is a client config,
		// and S3 Service will also terminate the idle connections, which we will rely on the retry
		// if two reads duration is too long and s3 service terminate the socket.
		request.WithResponseReadTimeout(10*time.Second))

	if getObjectErr != nil {
		return getObjectErr
	}
	r.resp = resp
	return nil
}

func (r *s3Reader) closeS3Socket() {
	if r.resp != nil {
		r.resp.Body.Close()
		r.resp = nil
	}
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

	//Close the socket when seeking
	//Special case: Dont close if the position hasn't changed
	if oldPos != r.offset {
		r.closeS3Socket()
	}

	return r.offset, nil
}

func min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}
