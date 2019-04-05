// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package file_system

//go:generate $MOCKGEN -destination=mock_file_system.go -package=file_system go.amzn.com/robomaker/bundle_support/file_system FileSystem
//go:generate $MOCKGEN -destination=mock_file.go -package=file_system go.amzn.com/robomaker/bundle_support/file_system File
//go:generate $MOCKGEN -destination=mock_file_info.go -package=file_system go.amzn.com/robomaker/bundle_support/file_system FileInfo

import (
	"io"
	"io/ioutil"
	"os"
	"time"
)

type FileSystem interface {
	NewFile(fd uintptr, name string) File
	Create(name string) (File, error)
	Open(name string) (File, error)
	Stat(name string) (FileInfo, error)
	RemoveAll(name string) error
	MkdirAll(name string, mode FileMode) error
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, mode FileMode) error
}

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	io.WriterAt
	Stat() (os.FileInfo, error)
}

type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() os.FileMode  // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() interface{}   // underlying data source (can return nil)
}

type FileMode os.FileMode

// osFS implements FileSystem using the local disk.
type osFS struct{}

func (osFS) NewFile(fd uintptr, name string) File      { return os.NewFile(fd, name) }
func (osFS) Create(name string) (File, error)          { return os.Create(name) }
func (osFS) Open(name string) (File, error)            { return os.Open(name) }
func (osFS) Stat(name string) (FileInfo, error)        { return os.Stat(name) }
func (osFS) RemoveAll(name string) error               { return os.RemoveAll(name) }
func (osFS) MkdirAll(name string, mode FileMode) error { return os.MkdirAll(name, os.FileMode(mode)) }
func (osFS) ReadFile(filename string) ([]byte, error)  { return ioutil.ReadFile(filename) }
func (osFS) WriteFile(filename string, data []byte, mode FileMode) error {
	return ioutil.WriteFile(filename, data, os.FileMode(mode))
}

func NewLocalFS() FileSystem {
	return &osFS{}
}
