// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

//go:generate $MOCKGEN -destination=mock_bundle_manager.go -package=bundle_support go.amzn.com/robomaker/bundle_support BundleProvider

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/archive"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/stream"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

/*
BundleProvider's responsibility is to create the Bundle object for the application to use.
It does this by:
1. Converting the URL to the bundle to a io.ReadSeeker
2. passing the ReadSeeker to the BundleArchive to know how extract the bundle
3. return a Bundle so that the application can use it.
*/
type BundleProvider interface {
	// With the URL to the bundle, extract bundle files to the cache, and return a useable Bundle
	GetBundle(url string) (Bundle, error)

	// Get the bundle from the URL, and expect the content id to match expectedContentId.
	// If it doesn't match, an error is thrown and we will get nil for bundle
	// NOTE: Any double quotes character in expectedContentId will be ignored.
	GetVersionedBundle(url string, expectedContentId string) (Bundle, error)

	// Set a progress callback function so that it will be called when we have a progress tick
	SetProgressCallback(callback ProgressCallback)

	// Set the rate at which the progress callback will be called.
	SetProgressCallbackRate(rateSeconds int)

	// Set a S3 client so we can use it to download from S3
	SetS3Client(s3Client s3iface.S3API)
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

type bundleProvider struct {
	s3Client                      s3iface.S3API
	bundleStore                   store.BundleStore
	progressCallback              ProgressCallback
	progressCallbackRateInSeconds int
}

func NewBundleProvider(bundleStore store.BundleStore) BundleProvider {
	return &bundleProvider{
		bundleStore:                   bundleStore,
		progressCallbackRateInSeconds: 1,
	}
}

func (b *bundleProvider) SetProgressCallback(callback ProgressCallback) {
	b.progressCallback = callback
}

func (b *bundleProvider) SetS3Client(s3Client s3iface.S3API) {
	b.s3Client = s3Client
}

func (b *bundleProvider) SetProgressCallbackRate(rateSeconds int) {
	b.progressCallbackRateInSeconds = rateSeconds
}

func (b *bundleProvider) GetVersionedBundle(url string, expectedContentId string) (Bundle, error) {
	return b.getBundle(url, expectedContentId)
}

func (b *bundleProvider) GetBundle(url string) (Bundle, error) {
	return b.getBundle(url, "")
}

func (b *bundleProvider) getBundle(url string, expectedContentId string) (Bundle, error) {
	// convert our URL to a readable seekable stream
	stream, contentId, contentLength, streamErr := stream.PathToStream(url, b.s3Client)
	if streamErr != nil {
		return nil, NewBundleError(streamErr, ERROR_TYPE_SOURCE)
	}

	if expectedContentId != "" && expectedContentId != contentId {
		return nil, NewBundleError(fmt.Errorf("Expected content ID [%v] does not match actual content ID [%v]", expectedContentId, contentId), ERROR_TYPE_CONTENT_ID)
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
	bundleArchive, bundleArchiveErr := archive.NewBundleArchive(stream)
	if bundleArchiveErr != nil {
		return nil, NewBundleError(bundleArchiveErr, ERROR_TYPE_FORMAT)
	}

	// ask our bundle archive to extract
	bundle, extractErr := bundleArchive.Extract(b.bundleStore)
	if extractErr != nil {
		return nil, NewBundleError(extractErr, ERROR_TYPE_EXTRACTION)
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
