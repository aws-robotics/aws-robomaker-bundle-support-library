// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Package bundle implements a provider to extract, parse, and manage bundles.

Basic Example:

	cachePath := "./cache"
	bundleStore := store.NewSimpleStore(cachePath)

	bundlePath := "bundle"
	prefixPath := "prefix"

	bundleProvider := bundle.NewProvider(bundleStore)
	b, _ := bundleProvider.GetBundle(bundlePath)
	b.SourceCommand()

*/
package bundle

import (
	"fmt"
	"path/filepath"
)

const (
	sourceCommandFormat              = "BUNDLE_CURRENT_PREFIX=%s source %s/setup.sh;"
	standardPosixSourceCommandFormat = "BUNDLE_CURRENT_PREFIX=%s . %s/setup.sh;"
)

// Bundle provides the source commands to
// apply the bundle's contents to the local environment
type Bundle interface {
	// List of commands that should be executed
	// to insert the bundle's contents into
	// a shell environment
	SourceCommands() []string

	// List of commands that should be executed
	// to insert the bundle's contents into
	// a POSIX-standard environment
	PosixSourceCommands() []string

	// List of commands that should be executed
	// to insert the bundle's contents into
	// a shell-standard environment
	//
	// The root path in the source commands will be replaced with location.
	// This is useful if you mount the store in a container want to access them in the container's mounted location.
	SourceCommandsUsingLocation(location string) []string

	// List of commands that should be executed
	// to insert the bundle's contents into
	// a POSIX-standard environment
	//
	// The root path in the source commands will be replaced with location.
	// This is useful if you mount the store in a container want to access them in the container's mounted location.
	PosixSourceCommandsUsingLocation(location string) []string

	// Releases all resources that this bundle holds
	Release()
}

// Create a new bundle. Give it an array of item paths. bundle knows how to construct source commands
// from the item paths
func newBundle(bundleStore Cache, itemKeys []string) Bundle {
	return &bundle{
		bundleStore: bundleStore,
		itemKeys:    itemKeys,
	}
}

// bundle's responsibility is to provide the correct source commands for the application to execute.
// The contents of the directories in the source commands are guaranteed to be available on disk by the BundleCache.
type bundle struct {
	bundleStore Cache
	itemKeys    []string
}

func (b *bundle) SourceCommands() []string {
	var sourceCommands []string
	bundleStorePath := b.bundleStore.RootPath()
	for _, itemKey := range b.itemKeys {
		itemPath := filepath.Join(bundleStorePath, itemKey)
		sourceCommand := fmt.Sprintf(sourceCommandFormat, itemPath, itemPath)
		sourceCommands = append(sourceCommands, sourceCommand)
	}
	return sourceCommands
}

func (b *bundle) PosixSourceCommands() []string {
	var sourceCommands []string
	bundleStorePath := b.bundleStore.RootPath()
	for _, itemKey := range b.itemKeys {
		itemPath := filepath.Join(bundleStorePath, itemKey)
		sourceCommand := fmt.Sprintf(standardPosixSourceCommandFormat, itemPath, itemPath)
		sourceCommands = append(sourceCommands, sourceCommand)
	}
	return sourceCommands
}

func (b *bundle) SourceCommandsUsingLocation(location string) []string {
	var sourceCommands []string
	for _, itemKey := range b.itemKeys {
		itemPath := filepath.Join(location, itemKey)
		sourceCommand := fmt.Sprintf(sourceCommandFormat, itemPath, itemPath)
		sourceCommands = append(sourceCommands, sourceCommand)
	}
	return sourceCommands
}

func (b *bundle) PosixSourceCommandsUsingLocation(location string) []string {
	var sourceCommands []string
	for _, itemKey := range b.itemKeys {
		itemPath := filepath.Join(location, itemKey)
		sourceCommand := fmt.Sprintf(standardPosixSourceCommandFormat, itemPath, itemPath)
		sourceCommands = append(sourceCommands, sourceCommand)
	}
	return sourceCommands
}

func (b *bundle) Release() {
	for _, itemKey := range b.itemKeys {
		b.bundleStore.Release(itemKey)
	}
}
