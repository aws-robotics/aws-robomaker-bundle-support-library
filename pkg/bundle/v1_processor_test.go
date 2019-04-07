// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"testing"
)

const (
	rootPath = "/testing_root"
)

// Matcher that tests for v1Extractor
type ofExtractorV1 struct {
}

func OfExtractorV1() gomock.Matcher {
	return &ofExtractorV1{}
}

func (o *ofExtractorV1) Matches(x interface{}) bool {
	_, ok := x.(*v1Extractor)
	return ok
}

func (o *ofExtractorV1) String() string {
	return "expected type: *extractors.v1Extractor"
}

func TestBundleProcessorV1_Extract_ShouldPutIntoStore(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	path := "testPath"

	mockBundleStore := NewMockCache(ctrl)
	mockBundleStore.EXPECT().Put(gomock.Any(), OfExtractorV1()).Return(path, nil)

	extractor := newBundleProcessorV1()
	bundle, err := extractor.extract(nil, mockBundleStore)

	assert.NotNil(t, bundle)
	assert.Nil(t, err)
}

func TestBundleProcessorV1_Extract_WithError_ShouldReturnError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	path := ""

	expectedError := fmt.Errorf("TestError")

	mockBundleStore := NewMockCache(ctrl)
	mockBundleStore.EXPECT().Put(gomock.Any(), OfExtractorV1()).Return(path, expectedError)

	extractor := newBundleProcessorV1()
	bundle, err := extractor.extract(nil, mockBundleStore)

	assert.Nil(t, bundle)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
}
