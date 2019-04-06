// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/extractors"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"github.com/google/uuid"
	"io"
)

func NewBundleProcessorV1() BundleProcessor {
	return &bundleProcessorV1{}
}

// Bundle v1 simply extracts tar.gz
type bundleProcessorV1 struct{}

func (b *bundleProcessorV1) Extract(inputStream io.ReadSeeker, bundleStore store.BundleStore) (Bundle, error) {
	// create a bundle extractor that knows how to extract the bundle
	bundleExtractor := extractors.NewBundleV1Extractor(inputStream)

	bundleKey := uuid.New().String()
	// put it into the store
	// for bundle v1, we plan to ask the higher-up caller for the key, use 12345 for now
	_, putErr := bundleStore.Put(bundleKey, bundleExtractor)
	if putErr != nil {
		return nil, putErr
	}
	return NewBundle(bundleStore, []string{bundleKey}), nil
}
