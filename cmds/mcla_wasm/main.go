
//go:build tinygo.wasm
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"syscall/js"

	. "github.com/kmcsr/mcla"
)

type Map = map[string]any

func getAPI()(m Map){
	return Map{
		"version": version,
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
		"parseLogErrors": js.FuncOf(func(_ js.Value, args []js.Value)(res any){
			buf, err := json.Marshal(parseLogErrors(args))
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
	api := getAPI()
	api["release"] = js.FuncOf(func(_ js.Value, _ []js.Value)(_ any){
		global.Delete("MCLA")
		close(exit)
		for _, v := range api {
			if fn, ok := v.(js.Func); ok {
				fn.Release()
			}
		}
		return js.Undefined()
	})
	global.Set("MCLA", api)
	fmt.Printf("MCLA-%s loaded\n", version)
	defer fmt.Printf("MCLA-%s unloaded\n", version)
	<-exit
}

func parseCrashReport(args []js.Value)(report *CrashReport){
	value := args[0]
	r := wrapJsValueAsReader(value)
	var err error
	if report, err = ParseCrashReport(r); err != nil {
		if err == io.EOF { // Couldn't find crash report, return null
			return nil
		}
		throw(err)
	}
	return
}

func parseLogErrors(args []js.Value)(errs []*JavaError){
	value := args[0]
	r := wrapJsValueAsReader(value)
	errs = ScanJavaErrors(r)
	return
}
