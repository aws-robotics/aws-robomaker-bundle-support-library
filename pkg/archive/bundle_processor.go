// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"io"
)

const (
	BundleProcessorVersion1 = "1"
	BundleProcessorVersion2 = "2"
)

// BundleProcessor's responsibility is to take a bundle stream and knows how to process/handle the Bundle File
// This includes the knowledge on how to process v1, v2, etc.
type BundleProcessor interface {

	// Extract takes the bundle bytes and extracts everything into the bundle store
	Extract(inputStream io.ReadSeeker, bundleCache store.BundleStore) (bundle.Bundle, error)
}

func BundleProcessorForVersion(version string) BundleProcessor {
	switch version {
	case BundleProcessorVersion1:
		return NewBundleProcessorV1()
	case BundleProcessorVersion2:
		return NewBundleProcessorV2()
	default:
		return nil
	}
}
