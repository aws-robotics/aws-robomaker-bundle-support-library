// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/s3"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

const (
	// Replace these values with actual location on S3

	// AWS S3 Bucket name where the bundle is stored.
	AwsS3BundleBucketName = "my-aws-bucket-name"

	// AWS S3 Key, includes the folder name and the file name of the bundle.
	// If the bundle file is directly in the bucket without sub-folders, just use the bundle file name.
	AwsS3BundleKey = "my-folder/my-bundle-file.tar"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))
	svc := awsS3.New(sess)
	s3Reader, err := s3.NewS3Reader(svc, AwsS3BundleBucketName, AwsS3BundleKey)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Etag is %v\n", s3Reader.Etag)

	//In this example, a small (1kb) intermediate buffer size is used.
	//Internally, the s3 reader will buffer 5MB at a time, so this won't call to s3 for each 1kb of data, which would be slow
	buffer := make([]byte, 1*1024)

	f, fileErr := os.Create("/tmp/my-bundle-file.tar")
	check(fileErr)
	for {
		bytesRead, readErr := s3Reader.Read(buffer)
		if readErr == io.EOF {
			_, writeErr := f.Write(buffer[:bytesRead])
			check(writeErr)
			f.Close()
			break
		}
		check(readErr)

		_, writeErr := f.Write(buffer[:bytesRead])
		check(writeErr)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
