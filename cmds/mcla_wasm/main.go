
//go:build tinygo.wasm
package main

import (
	"context"
	"fmt"
	"io"
	"syscall/js"

	. "github.com/kmcsr/mcla"
)

var bgCtx context.Context = createBackgroundCtx()
var releaseBgCtx context.CancelFunc

func createBackgroundCtx()(bgCtx context.Context){
	bgCtx, releaseBgCtx = context.WithCancel(context.Background())
	return
}

func getAPI()(m Map){
	return Map{
		"version": version,
		"parseCrashReport": asyncFuncOf(func(_ js.Value, args []js.Value)(res any){
			return asJsValue(parseCrashReport(args))
		}),
		"parseLogErrors": asyncFuncOf(func(_ js.Value, args []js.Value)(res any){
			return asJsValue(parseLogErrors(args))
		}),
		"analyzeLogErrors": asyncFuncOf(func(_ js.Value, args []js.Value)(res any){
			return analyzeLogErrors(args)
		}),
		"setGhDbPrefix": js.FuncOf(func(_ js.Value, args []js.Value)(res any){
			prefix := args[0]
			defaultErrDB.Prefix = prefix.String()
			return
		}),
	}
}

func main(){
	api := getAPI()
	api["release"] = js.FuncOf(func(_ js.Value, _ []js.Value)(_ any){
		global.Delete("MCLA")
		releaseBgCtx()
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
		panic(err)
	}
	return
}

func parseLogErrors(args []js.Value)(errs []*JavaError){
	value := args[0]
	r := wrapJsValueAsReader(value)
	errs = ScanJavaErrors(r)
	return
}

type logErrRes struct {
	Error     *JavaError            `json:"error"`
	Solutions []SolutionPossibility `json:"solutions"`
}

func analyzeLogErrors(args []js.Value)(result []logErrRes){
	value := args[0]
	r := wrapJsValueAsReader(value)
	errs := ScanJavaErrors(r)
	result = make([]logErrRes, len(errs))

	ctx, cancel := context.WithCancelCause(bgCtx)
	doneCh := make(chan struct{}, len(errs))
	for i, jerr := range errs {
		go func(){
			defer func(){
				doneCh <- struct{}{}
			}()
			var (
				res logErrRes
				err error
			)
			res.Error = jerr
			if res.Solutions, err = defaultAnalyzer.DoError(jerr); err != nil {
				cancel(err)
				return
			}
			result[i] = res
		}()
	}
	for i := 0; i < len(errs); i++ {
		select {
		case <-doneCh:
		case <-ctx.Done():
			panic(ctx.Err())
		}
	}
	return
}
