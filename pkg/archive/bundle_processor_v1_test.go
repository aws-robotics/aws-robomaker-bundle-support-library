// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/extractors"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"testing"
)

const (
	rootPath = "/testing_root"
)

// Matcher that tests for BundleV1Extractor
type ofExtractorV1 struct {
}

func OfExtractorV1() gomock.Matcher {
	return &ofExtractorV1{}
}

func (o *ofExtractorV1) Matches(x interface{}) bool {
	_, ok := x.(*extractors.BundleV1Extractor)
	return ok
}

func (o *ofExtractorV1) String() string {
	return "expected type: *extractors.BundleV1Extractor"
}

func TestBundleProcessorV1_Extract_ShouldPutIntoStore(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	path := "testPath"

	mockBundleStore := store.NewMockBundleStore(ctrl)
	mockBundleStore.EXPECT().Put(gomock.Any(), OfExtractorV1()).Return(path, nil)

	extractor := NewBundleProcessorV1()
	bundle, err := extractor.Extract(nil, mockBundleStore)

	assert.NotNil(t, bundle)
	assert.Nil(t, err)
}

func TestBundleProcessorV1_Extract_WithError_ShouldReturnError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	path := ""

	expectedError := fmt.Errorf("TestError")

	mockBundleStore := store.NewMockBundleStore(ctrl)
	mockBundleStore.EXPECT().Put(gomock.Any(), OfExtractorV1()).Return(path, expectedError)

	extractor := NewBundleProcessorV1()
	bundle, err := extractor.Extract(nil, mockBundleStore)

	assert.Nil(t, bundle)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
}
