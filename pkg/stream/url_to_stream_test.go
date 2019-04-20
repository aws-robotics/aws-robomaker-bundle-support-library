// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stream

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathToStream_WithUnsupportedUrl_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "https://test/stream"

	stream, _, _, err := URLToStream(filePath)

	assert.Nil(t, stream)
	assert.NotNil(t, err)
}
