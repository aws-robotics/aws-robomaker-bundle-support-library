## AWS Robomaker Bundle Support Library

**This API is currently under active development and should not be considered stable.**

A Library in Go that supports download and extraction of colcon-bundle format. https://github.com/colcon/colcon-bundle

## CLI

We provide a rudimentary CLI to expose the base functionality of this library. 
With `GOMODULE11=on` you can run it by executing:

`go run github.com/aws-robotics/aws-robomaker-bundle-support-library/cmd/cli`

Usage:

```
./cli --bundle my_bundle.tar

--bundle - Path to bundle file
--prefix - Prefix to put onto the source command. This is generally used when the CLI is run
on a host, but the source command will run inside a Docker container. If you have your cache 
directory mounted as '/cache' in the Docker container you should set prefix to '/cache'.

```

## License

This library is licensed under the Apache 2.0 License. 
