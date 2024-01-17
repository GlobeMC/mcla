package main

import (
	"syscall/js"
)

var (
	global = js.Global()
	// class
	Object                      = global.Get("Object")
	Reflect                     = global.Get("Reflect")
	Symbol                      = global.Get("Symbol")
	Promise                     = global.Get("Promise")
	Array                       = global.Get("Array")
	Uint8Array                  = global.Get("Uint8Array")
	ReadableStream              = global.Get("ReadableStream")
	ReadableStreamDefaultReader = global.Get("ReadableStreamDefaultReader")
	ReadableStreamBYOBReader    = global.Get("ReadableStreamBYOBReader")
	// function
	jsFetch = global.Get("fetch")
	// API
	console        = global.Get("console")
	caches         = global.Get("caches")
	sessionStorage = global.Get("sessionStorage")
	localStorage   = global.Get("localStorage")
)
