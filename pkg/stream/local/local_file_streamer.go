// Package local contains an implementation of
// streamer for the local file system
package local

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/fs"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/stream"
	"io"
	"regexp"
)

type streamer struct {
	fileSystem fs.FileSystem
}

// NewStreamer creates a new stream.streamer that can
// be used to stream from the local file system.
// It supports regular unix paths (`/regular/unix/paths`)
// along with `file://` URLs.
func NewStreamer() stream.Streamer {
	return &streamer{fs.NewLocalFS()}
}

func newStreamer(fileSystem fs.FileSystem) *streamer {
	return &streamer{fileSystem}
}

func (s *streamer) CanStream(url string) bool {
	_, err := parseURL(url)
	return err == nil
}

func (s *streamer) CreateStream(url string) (io.ReadSeeker, int64, string, error) {
	filePath, err := parseURL(url)
	if err != nil {
		return nil, 0, "", err
	}
	md5Sum, md5Err := md5SumFile(filePath, s.fileSystem)
	if md5Err != nil {
		return nil, 0, "", md5Err
	}

	file, openErr := s.fileSystem.Open(filePath)
	if openErr != nil {
		return nil, 0, "", openErr
	}

	fileInfo, statErr := file.Stat()
	if statErr != nil {
		return nil, 0, "", statErr
	}

	return file, fileInfo.Size(), md5Sum, nil

}

// Returns a normal filesystem path from a file url (file:///)
func parseURL(url string) (string, error) {
	r1 := regexp.MustCompile(`^(\/.*)`)
	r2 := regexp.MustCompile(`^file:\/\/(\/.*)`)
	result1 := r1.FindStringSubmatch(url)
	result2 := r2.FindStringSubmatch(url)
	if result1 != nil {
		return result1[1], nil
	} else if result2 != nil {
		return result2[1], nil
	}
	return "", fmt.Errorf("url: %v is not a valid file system url", url)
}

func md5SumFile(filePath string, fileSystem fs.FileSystem) (string, error) {
	file, openErr := fileSystem.Open(filePath)
	var hashString string
	if openErr != nil {
		return "", openErr
	}
	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return hashString, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	hashString = hex.EncodeToString(hashInBytes)
	return hashString, nil
}
