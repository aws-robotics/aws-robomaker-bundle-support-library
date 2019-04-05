// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

//go:generate $MOCKGEN -destination=mock_bundle.go -package=bundle go.amzn.com/robomaker/bundle_support/bundle Bundle

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"path/filepath"
)

const (
	sourceCommandFormat              = "BUNDLE_CURRENT_PREFIX=%s source %s/setup.sh;"
	standardPosixSourceCommandFormat = "BUNDLE_CURRENT_PREFIX=%s . %s/setup.sh;"
)

// Bundle's responsibility is to provide the correct source commands for the application to execute.
// The contents of the directories in the source commands are guaranteed to be available on disk by the BundleCache.
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
	// a shell/POSIX-standard environment
	//
	// The root path in the source commands will be replaced with location.
	// This is useful if you mount the store in a container want to access them in the container's mounted location.
	SourceCommandsUsingLocation(location string) []string
	PosixSourceCommandsUsingLocation(location string) []string

	// Releases all resources that this Bundle holds
	Release()
}

// Create a new bundle. Give it an array of item paths. Bundle knows how to construct source commands
// from the item paths
func NewBundle(bundleStore store.BundleStore, itemKeys []string) Bundle {
	return &bundle{
		bundleStore: bundleStore,
		itemKeys:    itemKeys,
	}
}

type bundle struct {
	bundleStore store.BundleStore
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
