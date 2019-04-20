// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

//go:generate mockgen -destination=mock_cache.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle Cache
//go:generate mockgen -destination=mock_extractor.go -package=bundle github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle Extractor

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/fs"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/stream"
	"time"
)

// Cache manages the contents of bundles in the local filesystem.
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

// Extractor extracts all its contents of an archive into the target location
type Extractor interface {
	// Extract contents to extractLocation using fs to write to the local file system.
	Extract(extractLocation string, fs fs.FileSystem) error
}

// ProgressCallback returns information about the download and extraction
// of the bundle to the caller.
type ProgressCallback func(percentDone float32, timeElapsed time.Duration)


// Provider accepts a URL pointing at a bundle and returns the corresponding
// bundle object. It supports fetching from any URL supported by URLToStream.
type Provider struct {
	bundleStore                   Cache
	progressCallback              ProgressCallback
	progressCallbackRateInSeconds int
}

// NewProvider creates a provider which uses the passed in Cache
// as storage for extracted bundles.
func NewProvider(bundleStore Cache) *Provider {
	return &Provider{
		bundleStore:                   bundleStore,
		progressCallbackRateInSeconds: 1,
	}
}

// SetProgressCallback accepts a function to be invoked
// at regular intervals during download and extraction.
func (b *Provider) SetProgressCallback(callback ProgressCallback) {
	b.progressCallback = callback
}

// SetProgressCallbackRate sets the rate in seconds the progress
// callback should be invoked.
func (b *Provider) SetProgressCallbackRate(rateSeconds int) {
	b.progressCallbackRateInSeconds = rateSeconds
}

// GetBundle fetches and extracts the bundle pointed to by url
// and returns its representation.
func (b *Provider) GetBundle(url string) (Bundle, error) {
	return b.GetVersionedBundle(url, "")
}

// GetVersionedBundle fetches and extracts the bundle pointed to by
// URL and verifies its hash matches the passed in expectedContentID.
// For S3 downloads the etag is used.
func (b *Provider) GetVersionedBundle(url string, expectedContentID string) (Bundle, error) {
	// convert our URL to a readable seekable stream
	stream, contentLength, contentID, streamErr := stream.URLToStream(url)
	if streamErr != nil {
		return nil, newBundleError(streamErr, errorTypeSource)
	}

	if expectedContentID != "" && expectedContentID != contentID {
		return nil, newBundleError(fmt.Errorf("Expected content ID [%v] does not match actual content ID [%v]", expectedContentID, contentID), errorTypeContentID)
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