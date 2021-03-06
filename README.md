[![Build Status](https://travis-ci.org/aws-robotics/aws-robomaker-bundle-support-library.svg?branch=development)](https://travis-ci.org/aws-robotics/aws-robomaker-bundle-support-library) [![Go Report Card](https://goreportcard.com/badge/github.com/aws-robotics/aws-robomaker-bundle-support-library)](https://goreportcard.com/report/github.com/aws-robotics/aws-robomaker-bundle-support-library)
[![GoDoc](https://godoc.org/github.com/aws-robotics/aws-robomaker-bundle-support-library?status.svg)](https://godoc.org/github.com/aws-robotics/aws-robomaker-bundle-support-library)
## AWS Robomaker Bundle Support Library

**This API is currently under active development and should not be considered stable.**

A Library in Go that supports download and extraction of colcon-bundle format. https://github.com/colcon/colcon-bundle

## CLI

We provide a rudimentary CLI to expose the base functionality of this library. 
With `GO111MODULE=on` you can run it by executing:

`go run github.com/aws-robotics/aws-robomaker-bundle-support-library/cmd/cli`

Usage:

```
./cli --bundle my_bundle.tar

--bundle - Path to bundle file
--prefix - Prefix to put onto the source command. This is generally used when the CLI is run
on a host, but the source command will run inside a Docker container. If you have your cache 
directory mounted as '/cache' in the Docker container you should set prefix to '/cache'.
--cache - Path to store extracted bundle contents (Default: ./cache)

```

## Developing

In order to build and run this package from source you should execute the following (Golang 1.16+ recommended):

```
source environment.sh
go get github.com/mitchellh/gox
go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen
go build ./...
go generate ./...
go test -v -race ./...
```

`environment.sh` is used to set `GOBIN` so that `mockgen` installs properly using Go modules.

NOTE: If you are using Golang version < 1.16 then you may need to set the environment variable `GO111MODULES=on`.

## License

This library is licensed under the Apache 2.0 License. 
