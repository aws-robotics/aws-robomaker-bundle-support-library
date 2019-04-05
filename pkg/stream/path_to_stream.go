// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stream

//go:generate $MOCKGEN -destination=mock_s3_client.go -package=stream github.com/aws/aws-sdk-go/service/s3/s3iface S3API

import (
	"crypto/md5"
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	s3Prefix      = "s3://"      //Looks like: s3://<bucket>/<key>
	s3HttpsPrefix = "https://s3" //Looks like: https://s3-<region>.amazonaws.com/<bucket>/<key>
	httpPrefix    = "http://"
	httpsPrefix   = "https://"
)

type md5Func func(filePath string, fileSystem file_system.FileSystem) (string, error)

// PathToStream converts a URL (currently supports: S3 URL and local file system)
// into a io.ReadSeeker for reading and seeking the contents
// return:
// 1. the io.ReadSeeker
// 2. a hash of the file -- for S3, etag. for local file, md5sum
// 3. the length of the stream in bytes
// 4. error if any
func PathToStream(url string, s3Client s3iface.S3API) (io.ReadSeeker, string, int64, error) {
	return pathToStream(url, s3Client, file_system.NewLocalFS(), md5SumFile)
}

func pathToStream(url string, s3Client s3iface.S3API, fileSystem file_system.FileSystem, computeMd5 md5Func) (io.ReadSeeker, string, int64, error) {
	// is this s3?
	if strings.HasPrefix(url, s3Prefix) || strings.HasPrefix(url, s3HttpsPrefix) {
		region, bucket, key, err := parseS3Url(url)
		if err != nil {
			return nil, "", 0, err
		}
		return pathToS3Stream(region, bucket, key, s3Client)
	}

	if strings.HasPrefix(url, httpsPrefix) || strings.HasPrefix(url, httpPrefix) {
		return nil, "", 0, fmt.Errorf("http/s url not yet supported")
	}

	// Assume this is local file
	return fileToStream(url, fileSystem, computeMd5)
}

func fileToStream(filePath string, fileSystem file_system.FileSystem, computeMd5 md5Func) (io.ReadSeeker, string, int64, error) {
	// compute md5sum
	md5Sum, md5Err := computeMd5(filePath, file_system.NewLocalFS())
	if md5Err != nil {
		return nil, "", 0, md5Err
	}

	file, openErr := fileSystem.Open(filePath)
	if openErr != nil {
		return nil, "", 0, openErr
	}

	fileInfo, statErr := file.Stat()
	if statErr != nil {
		return nil, "", 0, statErr
	}

	return file, md5Sum, fileInfo.Size(), nil
}

func pathToS3Stream(region, bucket, key string, s3Client s3iface.S3API) (io.ReadSeeker, string, int64, error) {
	var svc s3iface.S3API

	// if we're given a s3 client, use it
	if s3Client != nil {
		svc = s3Client
	} else {
		// else, try to create a s3 client
		if region == "" {
			region = os.Getenv("REGION")
		}

		if region == "" {
			return nil, "", 0, fmt.Errorf("Could not determine region for s3 bundle")
		}
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(region),
		}))
		svc = awsS3.New(sess)
	}

	s3Reader, err := s3.NewS3Reader(svc, bucket, key)
	if err != nil {
		return nil, "", 0, err
	}

	return s3Reader, s3Reader.Etag, s3Reader.ContentLength, nil
}

func md5SumFile(filePath string, fileSystem file_system.FileSystem) (string, error) {
	file, openErr := fileSystem.Open(filePath)
	if openErr != nil {
		return "", openErr
	}
	defer file.Close()

	hash := md5.New()

	_, hashErr := io.Copy(hash, file)
	if hashErr != nil {
		return "", hashErr
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func parseS3Url(s3Url string) (region string, bucket string, key string, err error) {
	r1 := regexp.MustCompile(`https://s3-(.*)\.amazonaws\.com/(.*?)/(.*)`)
	r2 := regexp.MustCompile(`s3://(.*?)/(.*)`)
	result1 := r1.FindStringSubmatch(s3Url)
	result2 := r2.FindStringSubmatch(s3Url)
	if result1 != nil {
		return result1[1], result1[2], result1[3], nil
	} else if result2 != nil {
		return "", result2[1], result2[2], nil
	}

	return "", "", "", fmt.Errorf("Url %v is not a valid s3 url", s3Url)
}
