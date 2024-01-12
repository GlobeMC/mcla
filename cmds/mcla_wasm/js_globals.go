package main

import (
	"encoding/json"
	"syscall/js"
)

var (
	global = js.Global()
	// class
	Object                      = global.Get("Object")
	Promise                     = global.Get("Promise")
	Uint8Array                  = global.Get("Uint8Array")
	ReadableStream              = global.Get("ReadableStream")
	ReadableStreamDefaultReader = global.Get("ReadableStreamDefaultReader")
	ReadableStreamBYOBReader    = global.Get("ReadableStreamBYOBReader")
	// function
	jsFetch = global.Get("fetch")
	// API
	caches         = global.Get("caches")
	sessionStorage = global.Get("sessionStorage")
	localStorage   = global.Get("localStorage")
)

const appStorageKeyPrefix = "com.github.kmcsr.mcla."

// TODO: use https://developer.mozilla.org/en-US/docs/Web/API/IDBFactory
func getStorageValue(key string, ptr any) (ok bool) {
	if !localStorage.Truthy() {
		return
	}
	key = appStorageKeyPrefix + key
	value, err := awaitPromise(localStorage.Call("getItem", key))
	if err != nil {
		return false
	}
	if !value.Truthy() {
		return false
	}
	if ptr != nil {
		if err := json.Unmarshal(([]byte)(value.String()), ptr); err != nil {
			panic(err)
		}
	}
	return true
}

func setStorageValue(key string, value any) {
	if !localStorage.Truthy() {
		return
	}
	key = appStorageKeyPrefix + key
	buf, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	localStorage.Call("setItem", key, (string)(buf))
}

func delStorageValue(key string) {
	if !localStorage.Truthy() {
		return
	}
	key = appStorageKeyPrefix + key
	localStorage.Call("removeItem", key)
}
