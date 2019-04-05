// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package store

//go:generate $MOCKGEN -destination=mock_bundle_store.go -package=store go.amzn.com/robomaker/bundle_support/store BundleStore

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/extractors"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"os"
	"path/filepath"
	"sync"
)

// BundleStore's responsibility is to manage bundle files on disk.
// Every entry in the storeItems is a directory containing extracted files that bundle uses.
// The key is a key that uniquely identifies the group of extracted files.
// Each entry can be a versioned sub-part or a versioned full part of a bundle.
type BundleStore interface {

	// Put a key into the storeItems, with it's corresponding contents.
	// If a key already exists, we ignore the Put command, in order to prevent double work.
	// A reader is passed in, so that BundleStore can write the contents, and extract to disk.
	// an extractor is passed in so that bundle can call the extractor to extract
	//
	// By default, a Put will set the item to "protected"
	//
	// Returns:
	// result for GetPath for this item that is put.
	// error if there are extract errors
	Put(key string, extractor extractors.Extractor) (string, error)

	// Load existing keys into memory from disk.
	// The initial key load into memory refcount is 0
	// Return:
	// error if key doesn't exist on disk
	Load(keys []string) error

	// Given a key, get the root path to the extracted files.
	// An empty string "" is returned if the key doesn't exist.
	GetPath(key string) string

	// Given a key, does it exist in the storeItems?
	Exists(key string) bool

	// Return the root path of the store
	RootPath() string

	// Get keys from in use (refcount > 0) storeItems
	GetInUseItemKeys() []string

	// Tell the store that we're done with this item
	Release(key string) error

	// Deletes storage space of items that are unreferenced
	Cleanup()
}

func NewSimpleStore(rootPath string, fileSystem file_system.FileSystem) BundleStore {
	return &simpleStore{
		rootPath:   rootPath,
		storeItems: make(map[string]storeItem),
		fileSystem: fileSystem,
	}
}

// Store Item records a key that has been put into the store.
// protected: boolean. Set to true when we first put into the storeItems. Able to set to false by API.
//            when cleaning up the storeItems, items with protected set to true will not be cleaned up.
type storeItem struct {
	key        string
	refCount   int
	pathToItem string
}

type simpleStore struct {
	rootPath   string
	storeItems map[string]storeItem
	fileSystem file_system.FileSystem
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

func (s *simpleStore) Put(key string, extractor extractors.Extractor) (string, error) {
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
