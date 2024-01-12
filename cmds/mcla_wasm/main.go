//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"io"
	"syscall/js"

	. "github.com/GlobeMC/mcla"
)

var bgCtx context.Context = createBackgroundCtx()
var releaseBgCtx context.CancelFunc

func createBackgroundCtx() (bgCtx context.Context) {
	bgCtx, releaseBgCtx = context.WithCancel(context.Background())
	return
}

func getAPI() (m Map) {
	return Map{
		"version": version,
		"parseCrashReport": asyncFuncOf(func(_ js.Value, args []js.Value) (res any) {
			return asJsValue(parseCrashReport(args))
		}),
		"parseLogErrors": asyncFuncOf(func(_ js.Value, args []js.Value) (res any) {
			return asJsValue(parseLogErrors(args))
		}),
		"analyzeLogErrors": asyncFuncOf(func(_ js.Value, args []js.Value) (res any) {
			return analyzeLogErrors(args)
		}),
		"analyzeLogErrorsIter": asyncFuncOf(func(_ js.Value, args []js.Value) (res any) {
			return analyzeLogErrorsIter(args)
		}),
		"setGhDbPrefix": js.FuncOf(func(_ js.Value, args []js.Value) (res any) {
			prefix := args[0]
			prefixStr := prefix.String()
			fmt.Printf("Set database as %q\n", prefixStr)
			ghRepoPrefix = prefixStr
			return
		}),
	}
}

func main() {
	defaultErrDB.RefreshCache()

	api := getAPI()
	api["release"] = js.FuncOf(func(_ js.Value, _ []js.Value) (_ any) {
		global.Delete("MCLA")
		releaseBgCtx()
		return js.Undefined()
	})
	global.Set("MCLA", api)

	fmt.Printf("MCLA-%s loaded\n", version)
	defer fmt.Printf("MCLA-%s unloaded\n", version)

	<-bgCtx.Done()
}

func parseCrashReport(args []js.Value) (report *CrashReport) {
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

func parseLogErrors(args []js.Value) (errs []*JavaError) {
	value := args[0]
	r := wrapJsValueAsReader(value)
	var err error
	if errs, err = ScanJavaErrors(r); err != nil {
		panic(err)
	}
	return
}

func analyzeLogErrors(args []js.Value) (result []*ErrorResult) {
	r := wrapJsValueAsReader(args[0])
	result = make([]*ErrorResult, 0, 5)
	resCh, ctx := defaultAnalyzer.DoLogStream(bgCtx, r)
	for {
		select {
		case res := <-resCh:
			result = append(result, res)
		case <-ctx.Done():
			panic(context.Cause(ctx))
		}
	}
	return
}

func analyzeLogErrorsIter(args []js.Value) (iterator js.Value) {
	r := wrapJsValueAsReader(args[0])
	result, ctx := defaultAnalyzer.DoLogStream(bgCtx, r)
	iterator = NewChannelIteratorContext(ctx, result)
	return
}
