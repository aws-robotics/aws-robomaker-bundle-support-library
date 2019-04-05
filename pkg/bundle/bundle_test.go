// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	rootPath          = "/testing_root"
	containerRootPath = "/container_root"
)

var itemKeys = []string{
	"item1",
	"item2",
	"item3",
}

var itemPaths = []string{
	"/testing_root/item1",
	"/testing_root/item2",
	"/testing_root/item3",
}

var containerItemPaths = []string{
	"/container_root/item1",
	"/container_root/item2",
	"/container_root/item3",
}

func TestBundle_SourceCommands_GivesExpectedCommands(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBundleStore := store.NewMockBundleStore(ctrl)
	mockBundleStore.EXPECT().RootPath().Return(rootPath).AnyTimes()

	bundle := NewBundle(mockBundleStore, itemKeys)
	sourceCommands := bundle.SourceCommands()

	assert.Equal(t, 3, len(sourceCommands))
	assert.Equal(t, fmt.Sprintf(sourceCommandFormat, itemPaths[0], itemPaths[0]), sourceCommands[0])
	assert.Equal(t, fmt.Sprintf(sourceCommandFormat, itemPaths[1], itemPaths[1]), sourceCommands[1])
	assert.Equal(t, fmt.Sprintf(sourceCommandFormat, itemPaths[2], itemPaths[2]), sourceCommands[2])

	posixCommands := bundle.PosixSourceCommands()
	assert.Equal(t, 3, len(posixCommands))
	assert.Equal(t, fmt.Sprintf(standardPosixSourceCommandFormat, itemPaths[0], itemPaths[0]), posixCommands[0])
	assert.Equal(t, fmt.Sprintf(standardPosixSourceCommandFormat, itemPaths[1], itemPaths[1]), posixCommands[1])
	assert.Equal(t, fmt.Sprintf(standardPosixSourceCommandFormat, itemPaths[2], itemPaths[2]), posixCommands[2])
}

func TestBundle_SourceCommandsUsingLocation_GivesExpectedCommands(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBundleStore := store.NewMockBundleStore(ctrl)

	bundle := NewBundle(mockBundleStore, itemKeys)
	sourceCommands := bundle.SourceCommandsUsingLocation(containerRootPath)

	assert.Equal(t, 3, len(sourceCommands))
	assert.Equal(t, fmt.Sprintf(sourceCommandFormat, containerItemPaths[0], containerItemPaths[0]), sourceCommands[0])
	assert.Equal(t, fmt.Sprintf(sourceCommandFormat, containerItemPaths[1], containerItemPaths[1]), sourceCommands[1])
	assert.Equal(t, fmt.Sprintf(sourceCommandFormat, containerItemPaths[2], containerItemPaths[2]), sourceCommands[2])

	posixCommands := bundle.PosixSourceCommandsUsingLocation(containerRootPath)
	assert.Equal(t, 3, len(posixCommands))
	assert.Equal(t, fmt.Sprintf(standardPosixSourceCommandFormat, containerItemPaths[0], containerItemPaths[0]), posixCommands[0])
	assert.Equal(t, fmt.Sprintf(standardPosixSourceCommandFormat, containerItemPaths[1], containerItemPaths[1]), posixCommands[1])
	assert.Equal(t, fmt.Sprintf(standardPosixSourceCommandFormat, containerItemPaths[2], containerItemPaths[2]), posixCommands[2])

}

func TestBundle_Release_ShouldReleaseItemKeys(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBundleStore := store.NewMockBundleStore(ctrl)
	mockBundleStore.EXPECT().Release(itemKeys[0])
	mockBundleStore.EXPECT().Release(itemKeys[1])
	mockBundleStore.EXPECT().Release(itemKeys[2])

	bundle := NewBundle(mockBundleStore, itemKeys)
	bundle.Release()
}
