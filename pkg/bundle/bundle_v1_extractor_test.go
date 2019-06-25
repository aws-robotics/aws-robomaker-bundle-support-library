// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsExpectedFile_WhenExpectedBundle_ShouldReturnTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, isExpectedFile("bundle.tar"))
}

func TestIsExpectedFile_WhenExpectedMetadata_ShouldReturnTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, isExpectedFile("metadata.tar"))
}

func TestIsExpectedFile_WhenNotExpected_ShouldReturnFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, isExpectedFile("unknown.txt"))
}
