// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"github.com/google/uuid"
	"io"
)

func newBundleProcessorV1() bundleProcessor {
	return &bundleProcessorV1{}
}

// bundle v1 simply extracts tar.gz
type bundleProcessorV1 struct{}

func (b *bundleProcessorV1) extract(inputStream io.ReadSeeker, bundleStore Cache) (Bundle, error) {
	// create a bundle extractor that knows how to Extract the bundle
	bundleExtractor := newBundleV1Extractor(inputStream)

	bundleKey := uuid.New().String()
	// put it into the store
	// for bundle v1, we plan to ask the higher-up caller for the key, use 12345 for now
	_, putErr := bundleStore.Put(bundleKey, bundleExtractor)
	if putErr != nil {
		return nil, putErr
	}
	return newBundle(bundleStore, []string{bundleKey}), nil
}
