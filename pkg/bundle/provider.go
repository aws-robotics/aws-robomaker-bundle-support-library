// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

//go:generate mockgen -destination=mock_cache.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle Cache
//go:generate mockgen -destination=mock_extractor.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle Extractor

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/stream"
	"io"
	"time"
)

// BundleCache's responsibility is to manage bundle files on disk.
// Every entry in the storeItems is a directory containing extracted files that bundle uses.
// The key is a key that uniquely identifies the group of extracted files.
// Each entry can be a versioned sub-part or a versioned full part of a bundle.
type Cache interface {

	// Put a key into the storeItems, with it's corresponding contents.
	// If a key already exists, we ignore the Put command, in order to prevent double work.
	// A reader is passed in, so that BundleCache can write the contents, and Extract to disk.
	// an extractor is passed in so that bundle can call the extractor to Extract
	//
	// By default, a Put will set the item to "protected"
	//
	// Returns:
	// result for GetPath for this item that is put.
	// error if there are Extract errors
	Put(key string, extractor Extractor) (string, error)

	// Load existing keys into memory from disk.
	// The initial key load into memory refcount is 0
	// Return:
	// error if key doesn't exist on disk
	Load(keys []string) error

	// Given a key, get the root path to the extracted files.
	// An empty string "" is returned if the key doesn't exist.
	GetPath(key string) string

	// Given a key, does it exist in the storeItems?
	Exists(key string) bool

	// Return the root path of the store
	RootPath() string

	// Get keys from in use (refcount > 0) storeItems
	GetInUseItemKeys() []string

	// Tell the store that we're done with this item
	Release(key string) error

	// Deletes storage space of items that are unreferenced
	Cleanup()
}

// Extractor's responsibility is to Extract all its contents into the target Extract location
type Extractor interface {
	Extract(extractLocation string, fs file_system.FileSystem) error
}

type ProgressCallback func(percentDone float32, timeElapsed time.Duration)

type proxyReadSeeker struct {
	r                     io.ReadSeeker
	contentLength         int64
	readStartTime         time.Time
	lastUpdated           time.Time
	callback              ProgressCallback
	callbackRateInSeconds int
}

// Provider's responsibility is to create the bundle object for the application to use.
// It supports fetching from any url supported by UrlToStream.
type Provider struct {
	bundleStore                   Cache
	progressCallback              ProgressCallback
	progressCallbackRateInSeconds int
}

func NewProvider(bundleStore Cache) *Provider {
	return &Provider{
		bundleStore:                   bundleStore,
		progressCallbackRateInSeconds: 1,
	}
}

func (b *Provider) SetProgressCallback(callback ProgressCallback) {
	b.progressCallback = callback
}

func (b *Provider) SetProgressCallbackRate(rateSeconds int) {
	b.progressCallbackRateInSeconds = rateSeconds
}

func (b *Provider) GetVersionedBundle(url string, expectedContentId string) (Bundle, error) {
	return b.getBundle(url, expectedContentId)
}

func (b *Provider) GetBundle(url string) (Bundle, error) {
	return b.getBundle(url, "")
}

func (b *Provider) getBundle(url string, expectedContentId string) (Bundle, error) {
	// convert our URL to a readable seekable stream
	stream, contentLength, contentId, streamErr := stream.UrlToStream(url)
	if streamErr != nil {
		return nil, newBundleError(streamErr, errorTypeSource)
	}

	if expectedContentId != "" && expectedContentId != contentId {
		return nil, newBundleError(fmt.Errorf("Expected content ID [%v] does not match actual content ID [%v]", expectedContentId, contentId), errorTypeContentId)
	}

	if b.progressCallback != nil {
		stream = &proxyReadSeeker{
			r:                     stream,
			contentLength:         contentLength,
			callback:              b.progressCallback,
			callbackRateInSeconds: b.progressCallbackRateInSeconds,
			readStartTime:         time.Now(),
			lastUpdated:           time.Now(),
		}
	}

	// create a bundle archive for the stream
	bundleArchive, bundleArchiveErr := newBundleArchive(stream)
	if bundleArchiveErr != nil {
		return nil, newBundleError(bundleArchiveErr, errorTypeFormat)
	}

	// ask our bundle archive to Extract
	bundle, extractErr := bundleArchive.Extract(b.bundleStore)
	if extractErr != nil {
		return nil, newBundleError(extractErr, errorTypeExtraction)
	}

	return bundle, nil
}

func (r *proxyReadSeeker) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)

	currentPos, _ := r.Seek(0, io.SeekCurrent)
	percentDone := float32((float64(currentPos) / float64(r.contentLength)) * 100)

	if time.Since(r.lastUpdated).Seconds() > float64(r.callbackRateInSeconds) || currentPos == r.contentLength {
		r.callback(percentDone, time.Since(r.readStartTime))
		r.lastUpdated = time.Now()
	}

	return
}

func (r *proxyReadSeeker) Seek(offset int64, whence int) (newOffset int64, err error) {
	if whence == io.SeekEnd {
		r.callback(100.0, time.Since(r.readStartTime))
	}

	return r.r.Seek(offset, whence)
}
