// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBundleProcessorForVersion_V1_ShouldReturnV1(t *testing.T) {
	t.Parallel()
	processor := processorForVersion(processorVersion1)

	// type assert that this is v1
	_, ok := processor.(*bundleProcessorV1)

	assert.NotNil(t, processor)
	assert.True(t, ok)
}

func TestBundleProcessorForVersion_V2_ShouldReturnV2(t *testing.T) {
	t.Parallel()
	processor := processorForVersion(processorVersion2)

	// type assert that this is v2
	_, ok := processor.(*bundleProcessorV2)

	assert.NotNil(t, processor)
	assert.True(t, ok)
}

func TestBundleProcessorForVersion_Unsupported_ShouldReturnNil(t *testing.T) {
	t.Parallel()
	processor := processorForVersion("NoVersion")

	assert.Nil(t, processor)
}
