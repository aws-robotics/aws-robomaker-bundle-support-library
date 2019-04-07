package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/bundle"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/file_system"
	"github.com/aws-robotics/aws-robomaker-bundle-support-library/pkg/store"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Bundle Helper"
	app.Usage = "Extracts a bundle and prints the command to source the bundle into a shell environment. " +
		"Will intelligently cache in the cache directory."
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "bundle", Value: "", Usage: "Path to bundle archive"},
		cli.StringFlag{Name: "prefix", Value: "", Usage: "Prefix to prepend onto the source command"},
		cli.StringFlag{Name: "cache", Value: "./cache", Usage: "Path to store extracted overlays"},
	}
	app.Action = func(c *cli.Context) error {
		cachePath := "./cache"
		bundleStore := store.NewSimpleStore(cachePath, file_system.NewLocalFS())

		bundlePath := c.String("bundle")
		prefixPath := c.String("prefix")

		if bundlePath == "" {
			err := errors.New("--bundle argument is required")
			log.Fatal(err)
			return err
		}

		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			err = os.Mkdir(cachePath, os.ModePerm)
			if err != nil {
				log.Fatal(err)
				return err
			}
		}

		files, err := ioutil.ReadDir(cachePath)
		if err != nil {
			log.Fatal(err)
			return err
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

		bundle_provider := bundle.NewBundleProvider(bundleStore)
		b, err := bundle_provider.GetBundle(bundlePath)
		if err != nil {
			log.Fatal(err)
			return err
		}
		for i := 0; i < len(b.PosixSourceCommands()); i++ {
			fmt.Print(b.PosixSourceCommandsUsingLocation(prefixPath)[i])
		}
		fmt.Print("\n")
		return nil
	}
	app.Run(os.Args)
}
