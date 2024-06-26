#!/usr/bin/env bash

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# go build -o ../run-server ./main
go build -C $THIS_DIR/src -o ../run-server ./main
