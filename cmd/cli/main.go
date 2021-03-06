// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Package cli provides a basic command line application to interact with bundles.

It will extract the bundle to the cache directory and then print the
command that can be used to source the bundle into the environment.

Usage:
  go run github.com/aws-robotics/aws-robomaker-bundle-support-library \
		--bundle <path to bundle> \
		--cache (optional) <path to cache directory (default: cache)> \
		--prefix (optional) <prefix for source command paths (must include cache directory)>
*/
package main

import (
	"errors"
	"fmt"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/stream/local"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/stream"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Bundle Helper"
	app.Usage = "Extracts a bundle and prints the command to source the bundle into a shell environment. " +
		"Will intelligently cache in the cache directory."
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "bundle", Value: "", Usage: "Path to bundle file"},
		cli.StringFlag{Name: "prefix", Value: "", Usage: "Prefix to put onto the source command"},
		cli.StringFlag{Name: "cache", Value: "cache", Usage: "Folder to be used as the cache " +
			"directory for extracted bundles."},
	}

	local := local.NewStreamer()
	stream.RegisterStreamer(local)

	app.Action = func(c *cli.Context) error {
		cachePath := c.String("cache")
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			err = os.Mkdir(cachePath, os.ModePerm)
			if err != nil {
				fmt.Printf("Failed to create cache directory: %s", cachePath)
				log.Fatal(err)
				return err
			}
		}
		bundleStore := store.NewSimpleStore(cachePath)

		bundlePath := c.String("bundle")
		if bundlePath == "" {
			fmt.Println("Bundle path cannot be empty.")
			return errors.New("bundle path cannot be empty")
		}
		absBundlePath, err := filepath.Abs(bundlePath)
		if err != nil {
			fmt.Printf("Bundle path is invalid: %s", bundlePath)
			log.Fatal(err)
			return err
		}
		prefixPath := c.String("prefix")

		files, err := ioutil.ReadDir(cachePath)
		if err != nil {
			log.Fatal(err)
		}

		var keys []string
		for _, file := range files {
			if file.IsDir() {
				keys = append(keys, file.Name())
			}
		}

		err = bundleStore.Load(keys)
		if err != nil {
			log.Fatal(err)
			return err
		}

		bundleProvider := bundle.NewProvider(bundleStore)
		b, err := bundleProvider.GetBundle(absBundlePath)
		if err != nil {
			log.Fatal(err)
			return err
		}
		for i := 0; i < len(b.PosixSourceCommands()); i++ {
			fmt.Println(b.PosixSourceCommandsUsingLocation(prefixPath)[i])
		}

		return nil
	}
	app.Run(os.Args)
}
