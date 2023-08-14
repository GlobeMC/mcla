
//go:build tinygo.wasm
package main

import (
	"io"
	"strings"
	"syscall/js"
)

type uint8ArrayReader struct {
	value js.Value
}

func (r uint8ArrayReader)ReadAt(buf []byte, offset int64)(n int, err error){
	sub := r.value.Call("subarray", (int)(offset), (int)(offset) + len(buf))
	n = js.CopyBytesToGo(buf, sub)
	if n != len(buf) {
		err = io.EOF
	}
	return
}

func wrapJsValueAsReader(value js.Value)(r io.Reader){
	switch value.Type() {
	case js.TypeString:
		return strings.NewReader(value.String())
	case js.TypeObject:
		if value.InstanceOf(Uint8Array) {
			return io.NewSectionReader(uint8ArrayReader{value}, 0, (1 <<  63) - 1)
		}
	}
	throw("Unexpect value type " + value.Type().String())
	return
}
