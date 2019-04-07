// Package local_file contains an implementation of Streamer for local filesystems.
package local

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"io"
	"regexp"
)

type Streamer struct {
	fileSystem file_system.FileSystem
}

// Creates a new stream.Streamer that can be used to stream from the local file system
func NewStreamer() *Streamer {
	return &Streamer{file_system.NewLocalFS()}
}

func newStreamer(fileSystem file_system.FileSystem) *Streamer {
	return &Streamer{fileSystem}
}

func (s *Streamer) CanStream(url string) bool {
	_, err := parseUrl(url)
	return err == nil
}

func (s *Streamer) CreateStream(url string) (io.ReadSeeker, int64, string, error) {
	filePath, err := parseUrl(url)
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
func parseUrl(url string) (string, error) {
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

func md5SumFile(filePath string, fileSystem file_system.FileSystem) (string, error) {
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