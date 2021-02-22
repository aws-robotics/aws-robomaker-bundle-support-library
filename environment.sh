#!/usr/bin/env sh

# Simple environment setup to work on a single go package
# https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md
# https://golang.org/cmd/go/#hdr-Environment_variables

export GOBIN=$PWD/bin
export PATH=$GOBIN:$PATH