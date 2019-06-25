// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"errors"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const (
	cacheRootPath                   = "/rootPath"
	sha256First                     = "1"
	sha256Second                    = "2"
	sha256Third                     = "3"
	expectedExtractLocationForFirst = "/rootPath/1"
)

func TestSimpleStore_Put_WithValidItem_ShouldPut(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExtractor := NewMockExtractor(ctrl)
	mockFileSystem := NewMockFileSystem(ctrl)

	// assert that this is called only once
	mockExtractor.EXPECT().Extract(expectedExtractLocationForFirst, mockFileSystem).Return(nil).Times(1)

	internalCache := make(map[string]storeItem)
	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
		fileSystem: mockFileSystem,
	}

	// assert doesn't exist in the storeItems
	assert.False(t, bundleStore.Exists(sha256First))

	// put
	putPath, putError := bundleStore.Put(sha256First, mockExtractor)

	item, exist := internalCache[sha256First]

	assert.Nil(t, putError)
	assert.True(t, bundleStore.Exists(sha256First))
	assert.Equal(t, expectedExtractLocationForFirst, putPath)
	assert.True(t, exist)
	assert.Equal(t, 1, item.refCount)

	// put a second time, and assert that no extraction is called
	secondPutPath, secondPutErr := bundleStore.Put(sha256First, mockExtractor)

	item2nd, exist := internalCache[sha256First]

	assert.Nil(t, secondPutErr)
	// assert it STILL exists
	assert.True(t, bundleStore.Exists(sha256First))
	assert.Equal(t, expectedExtractLocationForFirst, secondPutPath)
	assert.True(t, exist)
	assert.Equal(t, 2, item2nd.refCount)
}

func TestSimpleStore_Put_WithExtractError_ShouldNotPut(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExtractor := bundle.NewMockExtractor(ctrl)
	mockFileSystem := NewMockFileSystem(ctrl)

	extractError := errors.New("Extraction Error")

	// Return an extraction error
	mockExtractor.EXPECT().Extract(expectedExtractLocationForFirst, mockFileSystem).Return(extractError).Times(1)

	bundleStore := newSimpleStore(cacheRootPath, mockFileSystem)

	// assert doesn't exist in the storeItems
	assert.False(t, bundleStore.Exists(sha256First))

	// put
	putPath, putError := bundleStore.Put(sha256First, mockExtractor)

	assert.NotNil(t, putError)
	assert.False(t, bundleStore.Exists(sha256First))
	assert.Equal(t, "", putPath)
}

func TestSimpleStore_Load_WhenKeyDoesNotExist_ShouldLoad(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	mockFileSystem.EXPECT().Stat(gomock.Any()).Return(nil, nil).Times(1)

	internalCache := make(map[string]storeItem)
	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
		fileSystem: mockFileSystem,
	}
	// assert doesn't exist in the storeItems
	assert.False(t, bundleStore.Exists(sha256First))

	loadError := bundleStore.Load([]string{sha256First})

	assert.Nil(t, loadError)
	assert.True(t, bundleStore.Exists(sha256First))
	assert.Equal(t, 0, internalCache[sha256First].refCount)
}

func TestSimpleStore_Load_WhenKeyExist_ShouldLoad(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	mockFileSystem.EXPECT().Stat(gomock.Any()).Return(nil, nil).Times(3)

	internalCache := make(map[string]storeItem)
	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
		fileSystem: mockFileSystem,
	}

	bundleStore.Load([]string{sha256First, sha256Second, sha256Third})
	assert.True(t, bundleStore.Exists(sha256First))
	assert.True(t, bundleStore.Exists(sha256Second))
	assert.True(t, bundleStore.Exists(sha256Third))

	loadError := bundleStore.Load([]string{sha256First})
	assert.Nil(t, loadError)
}

func TestSimpleStore_Load_WhenKeyNotExistOnDisk_ShouldNotLoad(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	mockFileSystem.EXPECT().Stat(gomock.Any()).Return(nil, os.ErrNotExist).Times(1)

	internalCache := make(map[string]storeItem)
	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
		fileSystem: mockFileSystem,
	}

	loadError := bundleStore.Load([]string{sha256First})
	assert.NotNil(t, loadError)
}

func TestSimpleStore_Get_WhenDoesNotExist_ShouldReturnEmptyString(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	bundleStore := newSimpleStore(cacheRootPath, mockFileSystem)

	assert.Equal(t, "", bundleStore.GetPath(sha256First))
}

func TestSimpleStore_Get_WhenExist_ShouldReturnExpectedPath(t *testing.T) {
	t.Parallel()
	item := storeItem{
		key:        sha256First,
		refCount:   1,
		pathToItem: filepath.Join(cacheRootPath, sha256First),
	}
	internalCache := make(map[string]storeItem)
	internalCache[sha256First] = item

	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
	}

	assert.Equal(t, expectedExtractLocationForFirst, bundleStore.GetPath(sha256First))
}

func TestSimpleStore_GetInUseItemKeys_ShouldReturnInUseKeyOnly(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	itemNotInUse := storeItem{
		key:        sha256First,
		refCount:   0,
		pathToItem: filepath.Join(cacheRootPath, sha256First),
	}

	itemInUse := storeItem{
		key:        sha256Second,
		refCount:   1,
		pathToItem: filepath.Join(cacheRootPath, sha256Second),
	}

	internalCache := make(map[string]storeItem)
	internalCache[sha256First] = itemNotInUse
	internalCache[sha256Second] = itemInUse

	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
		fileSystem: mockFileSystem,
	}

	keysInUse := bundleStore.GetInUseItemKeys()

	assert.Contains(t, keysInUse, sha256Second)
	assert.NotContains(t, keysInUse, sha256First)
}

func TestSimpleStore_RootPath_ReturnsRootPath(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	rootPath := "testRootPath"
	bundleStore := newSimpleStore(rootPath, mockFileSystem)

	assert.Equal(t, rootPath, bundleStore.RootPath())
}

func TestSimpleStore_Release_KeyNotFound_ShouldReturnError(t *testing.T) {
	t.Parallel()
	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: make(map[string]storeItem),
	}

	err := bundleStore.Release(sha256First)

	assert.NotNil(t, err)
}

func TestSimpleStore_Release_KeyFound_ShouldRelease(t *testing.T) {
	t.Parallel()
	existingRefCount := 4

	item1 := storeItem{
		key:        sha256First,
		refCount:   existingRefCount,
		pathToItem: filepath.Join(cacheRootPath, sha256First),
	}
	internalCache := make(map[string]storeItem)
	internalCache[sha256First] = item1

	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
	}

	err := bundleStore.Release(sha256First)

	item, exists := internalCache[sha256First]

	assert.True(t, exists)
	assert.Equal(t, existingRefCount-1, item.refCount)
	assert.Nil(t, err)
}

func TestSimpleStore_Cleanup_ShouldCleanupOnlyUnprotectedItems(t *testing.T) {
	t.Parallel()
	item1 := storeItem{
		key:        sha256First,
		refCount:   1,
		pathToItem: filepath.Join(cacheRootPath, sha256First),
	}
	item2 := storeItem{
		key:        sha256Second,
		refCount:   0,
		pathToItem: filepath.Join(cacheRootPath, sha256Second),
	}
	item3 := storeItem{
		key:        sha256Third,
		refCount:   0,
		pathToItem: filepath.Join(cacheRootPath, sha256Third),
	}
	internalCache := make(map[string]storeItem)
	internalCache[sha256First] = item1
	internalCache[sha256Second] = item2
	internalCache[sha256Third] = item3

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileSystem := NewMockFileSystem(ctrl)

	mockFileSystem.EXPECT().RemoveAll(item2.pathToItem)
	mockFileSystem.EXPECT().RemoveAll(item3.pathToItem)

	bundleStore := simpleStore{
		rootPath:   cacheRootPath,
		storeItems: internalCache,
		fileSystem: mockFileSystem,
	}

	bundleStore.Cleanup()
}
