// Package s3 provides a Streamer implementation for AWS S3.
package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"io"
	"os"
	"regexp"
)

//go:generate mockgen -destination=mock_file_system.go -package=s3 github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system FileSystem
//go:generate mockgen -destination=mock_file.go -package=s3 github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system File
//go:generate mockgen -destination=mock_file_info.go -package=s3 github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system FileInfo

type Streamer struct {
	client s3iface.S3API
}

// Creates a new Streamer that can be used to stream from AWS S3 urls
// client can be nil and will then be created using the local environment
func NewStreamer(client s3iface.S3API) *Streamer {
	return &Streamer{client}
}

func (s *Streamer) CanStream(url string) bool {
	_, _, _, err := parseS3Url(url)
	return err == nil
}

func (s *Streamer) CreateStream(url string) (io.ReadSeeker, string, int64, error) {
	region, bucket, key, err := parseS3Url(url)
	if err != nil {
		return nil, "", 0, err
	}

	// Create a client if one was not provided
	if s.client == nil {
		s.client, err = createClientFromEnv(region)
		if err != nil {
			return nil, "", 0, err
		}
	}

	s3Reader, err := newS3ReaderBucketAndKey(s.client, bucket, key)
	if err != nil {
		return nil, "", 0, err
	}

	return s3Reader, s3Reader.Etag, s3Reader.ContentLength, nil
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

func createClientFromEnv(region string) (s3iface.S3API, error) {
	if region == "" {
		region = os.Getenv("REGION")
	}

	if region == "" {
		return nil, fmt.Errorf("Could not determine region for s3 bundle")
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	return s3.New(sess), nil
}
