
package main

import (
	"fmt"
	"syscall/js"
	"unsafe"
)

var (
	global = js.Global()
	Uint8Array = global.Get("Uint8Array")
)

//export throw
func throwStrPtr(str *byte, len int)

func throwString(str string){
	throwStrPtr(unsafe.StringData(str), len(str))
}

func throw(err any){
	switch e := err.(type) {
	case error:
		throwString(e.Error())
	case string:
		throwString(e)
	default:
		throwString(fmt.Sprint(err))
	}
}
