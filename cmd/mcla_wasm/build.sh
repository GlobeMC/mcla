#!/bin/sh

# exec tinygo build -target wasm "$@" "$(dirname $0)"
GOOS=js GOARCH=wasm exec go build "$@" "$(dirname $0)"
