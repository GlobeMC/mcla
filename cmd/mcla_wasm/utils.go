package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"syscall/js"
)

type Map = map[string]any

func asJsValue(v any) (res js.Value) {
	if v == nil {
		return js.Null()
	}
	if v0, ok := v.(js.Value); ok {
		return v0
	}
	if e, ok := v.(js.Error); ok {
		return e.Value
	}
	rv := reflect.ValueOf(v)
	switch rv.Type().Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return js.ValueOf(v)
	}
	buf, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error in asJsValue: json.Marshal:", err)
		panic(err)
	}
	var v1 any
	if err = json.Unmarshal(buf, &v1); err != nil {
		fmt.Println("Error in asJsValue: json.Unmarshal:", err)
		panic(err)
	}
	return js.ValueOf(v1)
}

var GoChannelIterator = (func() (cls js.Value) {
	cls = js.ValueOf(js.FuncOf(func(this js.Value, args []js.Value) (res any) {
		return
	}))
	cls.Set("name", "GoChannelIterator")
	cls.Set("length", 0)
	return
})()

var _emptyIterNextFn = (func() js.Func {
	cb := js.FuncOf(func(_ js.Value, args []js.Value) (res any) {
		resolve := args[0]
		resolve.Invoke(Map{"done": true, "value": nil})
		return
	})
	emptyIterNextPromise := Promise.New(cb)
	cb.Release()
	return js.FuncOf(func(_ js.Value, args []js.Value) (res any) {
		return emptyIterNextPromise
	})
})()

func NewChannelIteratorContext[T any](ctx context.Context, ch <-chan T) (iter js.Value) {
	iter = GoChannelIterator.New()
	var nextMethod, symAsyncItor js.Func
	nextMethod = asyncFuncOf(func(this js.Value, args []js.Value) (res any, err error) {
		select {
		case <-ctx.Done():
			iter.Set("next", _emptyIterNextFn)
			nextMethod.Release()
			symAsyncItor.Release()
			return nil, context.Cause(ctx)
		case val, ok := <-ch:
			if !ok {
				iter.Set("next", _emptyIterNextFn)
				nextMethod.Release()
				symAsyncItor.Release()
				return Map{"done": true, "value": nil}, nil
			}
			return Map{"done": false, "value": val}, nil
		}
	})
	symAsyncItor = js.FuncOf(func(this js.Value, args []js.Value) (res any) {
		return iter
	})
	iter.Set("next", nextMethod)
	Reflect.Call("set", iter, Symbol.Get("asyncIterator"), symAsyncItor)
	return
}

type readCloser struct {
	io.Reader
}

var _ io.ReadCloser = readCloser{}

func (r readCloser) Close() (err error) {
	if c, ok := r.Reader.(io.Closer); ok {
		return c.Close()
	}
	return
}

type (
	uint8ArrayReader struct {
		value js.Value
	}
	readableStreamDefaultReaderWrapper struct {
		off   int
		buf   *uint8ArrayReader
		value js.Value
	}
	readableStreamBYOBReaderWrapper struct {
		value js.Value
	}
)

var _ io.ReaderAt = uint8ArrayReader{}
var _ io.ReadCloser = readableStreamDefaultReaderWrapper{}

// var _ io.Reader = readableStreamBYOBReaderWrapper{} // TODO if necessary

func (r uint8ArrayReader) ReadAt(buf []byte, offset int64) (n int, err error) {
	sub := r.value.Call("subarray", (int)(offset), (int)(offset)+len(buf))
	n = js.CopyBytesToGo(buf, sub)
	if n != len(buf) {
		err = io.EOF
	}
	return
}

func (r readableStreamDefaultReaderWrapper) readFromInternalBuf(buf []byte) (n int, err error) {
	if r.buf != nil {
		n, err = r.buf.ReadAt(buf, (int64)(r.off))
		r.off += n
		if err == io.EOF {
			r.off = 0
			r.buf = nil
		}
	}
	return
}

func (r readableStreamDefaultReaderWrapper) Read(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return
	}
	if n, err = r.readFromInternalBuf(buf); n != 0 || err != nil {
		return
	}
	res, err := awaitPromise(r.value.Call("read"))
	if err != nil {
		return
	}
	if res.Get("done").Bool() {
		return 0, io.EOF
	}
	r.buf = &uint8ArrayReader{res.Get("value")}
	return r.readFromInternalBuf(buf)
}

func (r readableStreamDefaultReaderWrapper) Close() (err error) {
	awaitPromise(r.value.Call("cancel"))
	r.value.Call("releaseLock")
	return
}

func wrapJsValueAsReader(value js.Value) (r io.Reader, err error) {
	switch value.Type() {
	case js.TypeString:
		return strings.NewReader(value.String()), nil
	case js.TypeObject:
		if value.InstanceOf(Uint8Array) {
			return io.NewSectionReader(uint8ArrayReader{value}, 0, (1<<63)-1), nil
		}
		if value.InstanceOf(ReadableStream) {
			value = value.Call("getReader" /*, Map{ "mode": "byob" } TODO*/)
		}
		if value.InstanceOf(ReadableStreamDefaultReader) {
			return readableStreamDefaultReaderWrapper{value: value}, nil
		}
		// if value.InstanceOf(ReadableStreamBYOBReader) { // TODO
		// 	return readableStreamBYOBReaderWrapper{ value }
		// }
	}
	return nil, fmt.Errorf("Unexpect value type %s", value.Type())
}

// have to ensure the argument is a really JS Promise instance
func wrapPromise(promise js.Value) (done <-chan js.Value, err <-chan error) {
	done0 := make(chan js.Value, 1)
	err0 := make(chan error, 1)

	var success, failed js.Func
	success = js.FuncOf(func(_ js.Value, args []js.Value) (res any) {
		success.Release()
		failed.Release()
		done0 <- args[0]
		return
	})
	failed = js.FuncOf(func(_ js.Value, args []js.Value) (res any) {
		success.Release()
		failed.Release()
		err0 <- js.Error{args[0]}
		return
	})
	promise.Call("then", success).Call("catch", failed)
	return done0, err0
}

func awaitPromiseContext(ctx context.Context, promise js.Value) (res js.Value, err error) {
	if promise.Type() != js.TypeObject || !promise.InstanceOf(Promise) {
		return promise, nil
	}
	done, errCh := wrapPromise(promise)
	select {
	case res = <-done:
		return
	case err = <-errCh:
		return
	case <-ctx.Done():
		err = context.Cause(ctx)
		return
	}
}

func awaitPromise(promise js.Value) (res js.Value, err error) {
	return awaitPromiseContext(bgCtx, promise)
}

type asyncFuncSignature = func(this js.Value, args []js.Value) (res any, err error)

func asyncFuncOf(fn asyncFuncSignature) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) (res any) {
		var resolve, reject js.Value
		pcb := js.FuncOf(func(_ js.Value, args []js.Value) (res any) {
			resolve, reject = args[0], args[1]
			return
		})
		res = Promise.New(pcb)
		pcb.Release()
		go func() {
			defer func() {
				if e := recover(); e != nil {
					if je, ok := e.(js.Error); ok {
						reject.Invoke(je.Value)
					} else if er, ok := e.(error); ok {
						reject.Invoke(er.Error())
					} else {
						reject.Invoke(asJsValue(e))
					}
				}
			}()
			if res, err := fn(this, args); err != nil {
				reject.Invoke(err.Error())
			} else {
				resolve.Invoke(asJsValue(res))
			}
		}()
		return
	})
}

func foreachJsIterator(iterator js.Value, callback func(js.Value) error) (err error) {
	for {
		res := iterator.Call("next")
		if !res.Get("done").Bool() {
			break
		}
		if err = callback(res.Get("value")); err != nil {
			return
		}
	}
	return
}
