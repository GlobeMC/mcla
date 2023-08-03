
//go:build tinygo.wasm
package main

import (
	"encoding/json"
	"syscall/js"

	. "github.com/kmcsr/mcla"
)

var (
	global = js.Global()
	Uint8Array = global.Get("Uint8Array")
)

type Map = map[string]any

func getAPI()(m Map){
	return Map{
		"parseCrashReport": js.FuncOf(func(_ js.Value, args []js.Value)(res any){
			buf, err := json.Marshal(parseCrashReport(args))
			if err != nil {
				throw(err)
			}
			if err = json.Unmarshal(buf, &res); err != nil {
				throw(err)
			}
			return
		}),
	}
}

func main(){
	exit := make(chan struct{}, 0)
	global.Set("MCLA", getAPI())
	println("MCLA v?.?.? loaded")
	<-exit
}

func parseCrashReport(args []js.Value)(report *CrashReport){
	value := args[0]
	r := wrapJsValueAsReader(value)
	var err error
	if report, err = ParseCrashReport(r); err != nil {
		throw(err)
	}
	return
}
