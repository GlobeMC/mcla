
//go:build tinygo.wasm
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"syscall/js"

	. "github.com/kmcsr/mcla"
)

var bgCtx context.Context

type Map = map[string]any

func asJsValue(v any)(res any){
	buf, err := json.Marshal(v)
	if err != nil {
		throw(err)
	}
	if err = json.Unmarshal(buf, &res); err != nil {
		throw(err)
	}
	return
}

func getAPI()(m Map){
	return Map{
		"version": version,
		"parseCrashReport": js.FuncOf(func(_ js.Value, args []js.Value)(res any){
			return asJsValue(parseCrashReport(args))
		}),
		"parseLogErrors": js.FuncOf(func(_ js.Value, args []js.Value)(res any){
			return asJsValue(parseLogErrors(args))
		}),
	}
}

func main(){
	var release context.CancelFunc
	bgCtx, release = context.WithCancel(context.Background())

	api := getAPI()
	api["release"] = js.FuncOf(func(_ js.Value, _ []js.Value)(_ any){
		global.Delete("MCLA")
		release()
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

	<-bgCtx.Done()
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
