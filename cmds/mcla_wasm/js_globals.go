
package main

import (
	"encoding/json"
	"syscall/js"
)

var (
	global = js.Global()
	// class
	Object         = global.Get("Object")
	Promise        = global.Get("Promise")
	Uint8Array     = global.Get("Uint8Array")
	ReadableStream = global.Get("ReadableStream")
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

func getStorageValue(key string, ptr any)(ok bool){
	key = appStorageKeyPrefix + key
	value := localStorage.Call("getItem", key)
	if js.Null().Equal(value) {
		return false
	}
	if ptr != nil {
		if err := json.Unmarshal(([]byte)(value.String()), ptr); err != nil {
			panic(err)
		}
	}
	return true
}

func setStorageValue(key string, value any){
	key = appStorageKeyPrefix + key
	buf, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	localStorage.Call("setItem", key, (string)(buf))
}

func delStorageValue(key string){
	key = appStorageKeyPrefix + key
	localStorage.Call("removeItem", key)
}
