// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package store provides a simple implementation of the bundle.Cache interface
package store

//go:generate mockgen -destination=mock_file_system.go -package=store github.com/spf13/afero File
//go:generate mockgen -destination=mock_file.go -package=store github.com/spf13/afero Fs
//go:generate mockgen -destination=mock_file_info.go -package=store os FileInfo
//go:generate mockgen -destination=mock_extractor.go -package=store github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle Extractor

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"sync"
)

// NewSimpleStore returns a new bundle.Cache to provide
// caching for a bundle provider rootpath is the root
// directory you want the cache to use for storage
func NewSimpleStore(rootPath string) bundle.Cache {
	return &simpleStore{
		rootPath:   rootPath,
		storeItems: make(map[string]storeItem),
		fileSystem: afero.NewOsFs(),
	}
}

func newSimpleStore(rootPath string, fileSystem afero.Fs) bundle.Cache {
	return &simpleStore{
		rootPath:   rootPath,
		storeItems: make(map[string]storeItem),
		fileSystem: fileSystem,
	}
}

// Store Item records a key that has been put into the store.
// protected: boolean. Set to true when we first put into the storeItems. Able to set to false by API.
// when cleaning up the storeItems, items with protected set to true will not be cleaned up.
type storeItem struct {
	key        string
	refCount   int
	pathToItem string
}

type simpleStore struct {
	rootPath   string
	storeItems map[string]storeItem
	fileSystem afero.Fs
	mutex      sync.Mutex
}

func (s *simpleStore) Load(keys []string) error {
	// ensure that Load is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, key := range keys {
		if _, exists := s.storeItems[key]; exists {
			continue
		}

		itemPath, err := s.getPathToItemAndExistCheck(key)

		if err != nil {
			return err
		}

		// create a storeItem from this, 0 refcount from initial
		newItem := storeItem{
			key:        key,
			refCount:   0,
			pathToItem: itemPath,
		}
		s.storeItems[key] = newItem
	}
	return nil
}

func (s *simpleStore) Put(key string, extractor bundle.Extractor) (string, error) {
	// ensure that Put is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// there already exists an item, don't extract
	if item, exists := s.storeItems[key]; exists {
		// increment the item's refcount
		item.refCount++
		s.storeItems[key] = item
		return item.pathToItem, nil
	}

	// figure location to extract the files to and make the dir
	itemPath := s.getPathToItem(key)

	// create a storeItem from this
	newItem := storeItem{
		key:        key,
		refCount:   1,
		pathToItem: itemPath,
	}

	// now try to extract to the destination path
	extractErr := extractor.Extract(itemPath, s.fileSystem)
	if extractErr != nil {
		return "", extractErr
	}

	// no error, let's add it to the storeItems
	s.storeItems[key] = newItem
	return itemPath, nil
}

func (s *simpleStore) GetPath(key string) string {
	// ensure that GetPath is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	item, exists := s.storeItems[key]

	if !exists {
		return ""
	}

	return item.pathToItem
}

func (s *simpleStore) Exists(key string) bool {
	// ensure that Exists is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.storeItems[key]
	return exists
}

func (s *simpleStore) RootPath() string {
	return s.rootPath
}

// Internally, since we're using refCount, Release will decrement the refCount by 1
func (s *simpleStore) Release(key string) error {
	// ensure that Release is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	item, exists := s.storeItems[key]

	if !exists {
		return fmt.Errorf("key: %s does not exist", key)
	}

	item.refCount--

	// put the item back into the storeItems
	s.storeItems[key] = item
	return nil
}

func (s *simpleStore) Cleanup() {
	// ensure that Cleanup is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// iterate all keys in our map, and only delete unprotected
	var unreferencedItems []storeItem
	for _, item := range s.storeItems {
		if item.refCount < 1 {
			unreferencedItems = append(unreferencedItems, item)
		}
	}

	// now, delete the unprotected items
	for _, item := range unreferencedItems {
		s.fileSystem.RemoveAll(item.pathToItem)
		delete(s.storeItems, item.key)
	}
}

func (s *simpleStore) GetInUseItemKeys() []string {
	// ensure that GetInUseItemKeys is an atomic operation
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var inUseKeys []string
	for _, item := range s.storeItems {
		if item.refCount > 0 {
			inUseKeys = append(inUseKeys, item.key)
		}
	}

	return inUseKeys
}

func (s *simpleStore) getPathToItem(itemKey string) string {
	return filepath.Join(s.rootPath, itemKey)
}

func (s *simpleStore) getPathToItemAndExistCheck(itemKey string) (string, error) {

	itemPath := filepath.Join(s.rootPath, itemKey)
	if _, err := s.fileSystem.Stat(itemPath); os.IsNotExist(err) {
		return "", fmt.Errorf("itemPath: %s does not exist", itemPath)
	}
	return itemPath, nil
}
