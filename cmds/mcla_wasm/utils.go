
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

var _ io.ReaderAt = uint8ArrayReader{}

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

func lcsSplit[T comparable](a, b []T)(n int, a1, a2, b1, b2 []T){
	if len(a) == 0 || len(b) == 0 {
		return 0, a, nil, b, nil
	}
	type ele struct {
		a, b int
		len int
	}
	ch := make([]ele, len(b) + 1)
	for i, p := range a {
		var last ele
		for j, q := range b {
			cur := ch[j+1]
			if p == q {
				if last.len == 0 {
					last = ele{i, j, 1}
				}else if last.a + last.len == i && last.b + last.len == j {
					last.len++
				}
				ch[j + 1] = last
			}else if prev := ch[j]; prev.len > cur.len {
				ch[j + 1] = prev
			}
			last = cur
		}
	}
	res := ch[len(b)]
	if res.len == 0 {
		return
	}
	return res.len, a[:res.a], a[res.a + res.len:], b[:res.b], b[res.b + res.len:]
}

func lcsLength[T comparable](a, b []T)(n int){
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n, a1, a2, b1, b2 := lcsSplit(a, b)
	n += lcsLength(a1, b1) + lcsLength(a2, b2)
	return
}

func lcsPercent[T comparable](a, b []T)(v float32){
	if len(b) > len(a) {
		a, b = b, a
	}
	if len(b) == 0 {
		return 0.0
	}
	n := (float32)(lcsLength(a, b))
	return n / (float32)(len(a))
}

func splitWords(line string)(words []string){
	words = strings.Fields(line)
	for i, w := range words {
		words[i] = strings.Trim(w, ",.") // remove punctuation marks
	}
	return
}

func lineSamePercent(a, b string)(float32){
	return lcsPercent(splitWords(a), splitWords(b))
}
