#!/bin/sh

exec tinygo build -target wasm "$@" "$(dirname $0)"
