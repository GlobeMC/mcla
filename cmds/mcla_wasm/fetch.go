package main

import (
	"context"
	"io"
	"net/textproto"
	"syscall/js"
)

type Header = textproto.MIMEHeader

type Response struct {
	native     js.Value
	Status     string
	StatusCode int
	Body       io.ReadCloser
	Type       string
	Url        string
	Header     Header
}

func fetchContext(ctx context.Context, url string, opts ...Map) (res *Response, err error) {
	args := make([]any, 1, 2)
	args[0] = url
	if len(opts) > 0 {
		args = append(args, opts)
	}
	var res0 js.Value
	if res0, err = awaitPromiseContext(ctx, jsFetch.Invoke(args...)); err != nil {
		return
	}
	header0 := res0.Get("headers")
	header := make(Header, 10)
	foreachJsIterator(header0.Call("entries"), func(v js.Value) (err error) {
		header.Set(v.Index(0).String(), v.Index(1).String())
		return
	})
	res = &Response{
		native:     res0,
		Status:     res0.Get("statusText").String(),
		StatusCode: res0.Get("status").Int(),
		Body:       readCloser{wrapJsValueAsReader(res0.Get("body"))},
		Type:       res0.Get("type").String(),
		Url:        res0.Get("url").String(),
		Header:     header,
	}
	return
}

func fetch(url string, opts ...Map) (res *Response, err error) {
	return fetchContext(bgCtx, url, opts...)
}
