#!/bin/sh

exec tinygo build -target wasm -opt=z -no-debug "$@" "$(dirname $0)"
