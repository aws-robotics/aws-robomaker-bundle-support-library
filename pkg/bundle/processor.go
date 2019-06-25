// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"io"
)

const (
	processorVersion1 = "1"
	processorVersion2 = "2"
)

// bundleProcessor's responsibility is to take a bundle stream and knows how to process/handle the bundle File
// This includes the knowledge on how to process v1, v2, etc.
type bundleProcessor interface {
	// Extract takes the bundle bytes and extracts everything into the bundle store
	extract(inputStream io.ReadSeeker, bundleCache Cache) (Bundle, error)
}

func processorForVersion(version string) bundleProcessor {
	switch version {
	case processorVersion1:
		return newBundleProcessorV1()
	case processorVersion2:
		return newBundleProcessorV2()
	default:
		return nil
	}
}
