// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"time"
)

const (
	// Replace these values with actual location on your disk

	// BundleStoreLocation is the root directory on local disk where bundles will extract to
	BundleStoreRootLocation = "/my-bundle-store-root"

	// BundleFileLocation is the location of your bundle file. This can be local disk or a S3 URL.
	BundleFileLocation = "/my-bundleFile"
)

func main() {
	// Create a bundle store pointing to a local directory
	localFileSystem := file_system.NewLocalFS()
	bundleStore := store.NewSimpleStore(BundleStoreRootLocation, localFileSystem)

	bundleProvider := bundle.NewBundleProvider(bundleStore)
	bundleProvider.SetProgressCallback(printProgress)

	// Start measuring the time it takes to extract
	startTime := time.Now()

	// Call the Bundle Provider to extract and get the bundle.
	bundle, bundleErr := bundleProvider.GetBundle(BundleFileLocation)

	if bundleErr != nil {
		fmt.Printf("Get Bundle Error: %v\n", bundleErr)
		return
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("Time taken to download and extract: %s\n", elapsedTime)
	fmt.Printf("SourceCommands: %v\n", bundle.SourceCommands())

	// Release will mark the current bundle as unused, but will not be deleted.
	bundle.Release()

	// Now extract a second time, because everything is in the BundleStore, this will only take a short amount of time
	startTime = time.Now()
	bundle, bundleErr = bundleProvider.GetBundle(BundleFileLocation)
	if bundleErr != nil {
		fmt.Printf("Get Bundle Error: %v\n", bundleErr)
		return
	}

	elapsedTime = time.Since(startTime)
	fmt.Printf("Time taken to download and extract: %s\n", elapsedTime)
	fmt.Printf("SourceCommands: %v\n", bundle.SourceCommandsUsingLocation("/abc"))

	bundle.Release()

	// Cleanup will delete all bundles that are marked unused.
	bundleStore.Cleanup()
}

func printProgress(percentDone float32, timeElapsed time.Duration) {
	percentLeft := 100.0 - percentDone

	//This is only "accurate" as long as download speed doesn't fluctuate too much. For better results, the estimate
	//should be based on the time taken for the last few percent done, not the entire percent done
	estimatedTimeRemaining := time.Duration((float32(timeElapsed.Seconds())/percentDone)*percentLeft) * time.Second

	//Round to the nearest second to avoid long strings like 1.794995859s
	timeElapsedRounded := time.Duration(timeElapsed.Nanoseconds()/time.Second.Nanoseconds()) * time.Second
	estimatedTimeRemainingRounded := time.Duration(estimatedTimeRemaining.Nanoseconds()/time.Second.Nanoseconds()) * time.Second

	//Don't show time left for the first few percent, it'll be inaccurate since we are using a rolling average
	if percentDone < 3.0 {
		fmt.Printf("Percent done: %.2f. Time elapsed: %v. Estimated time remaining: --\n", percentDone, timeElapsedRounded)
	} else {
		fmt.Printf("Percent done: %.2f. Time elapsed: %v. Estimated time remaining: %v\n", percentDone, timeElapsedRounded, estimatedTimeRemainingRounded)
	}
}
