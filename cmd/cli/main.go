package main

import (
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
	app.Name = "Bundle Info"
	app.Usage = "Get source command for bundle"
	app.Flags = []cli.Flag {
		cli.StringFlag{Name: "bundle", Value: "", Usage: "Path to bundle file"},
		cli.StringFlag{Name: "prefix", Value: "", Usage: "Prefix to put onto the source command"},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("Opening ", c.String("bundle"))
		cache_path := "./cache"
		bundle_store := store.NewSimpleStore(cache_path, file_system.NewLocalFS())

		bundle_path := c.String("bundle")
		prefix_path := c.String("prefix")

		files, err := ioutil.ReadDir(cache_path)
		if err != nil {
			log.Fatal(err)
		}

		var keys []string
		for _, file := range files {
			if file.IsDir() {
				keys = append(keys, file.Name())
			}
		}

		err = bundle_store.Load(keys)
		if err != nil {
			log.Fatal(err)
			return err
		}

		bundle_provider := bundle.NewBundleProvider(bundle_store)
		b, err := bundle_provider.GetBundle(bundle_path)
		if err != nil {
			log.Fatal(err)
			return err
		}
		for i := 0; i < len(b.PosixSourceCommands()); i++  {
			fmt.Print(b.PosixSourceCommandsUsingLocation(prefix_path)[i])
		}

		return nil
	}
	app.Run(os.Args)
}



