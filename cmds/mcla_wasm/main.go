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
		"parseCrashReport": asyncFuncOf(func(_ js.Value, args []js.Value) (res any, err error) {
			return parseCrashReport(args)
		}),
		"parseLogErrors": asyncFuncOf(func(_ js.Value, args []js.Value) (res any, err error) {
			return parseLogErrors(args)
		}),
		"analyzeLogErrors": asyncFuncOf(func(_ js.Value, args []js.Value) (res any, err error) {
			return analyzeLogErrors(args)
		}),
		"analyzeLogErrorsIter": asyncFuncOf(func(_ js.Value, args []js.Value) (res any, err error) {
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

const LICENSE = `mcla (Minecraft Log Analyzer) %s
Copyright (C) 2023 <zyxkad@gmail.com> all rights reserved
Under GNU AFFERO GENERAL PUBLIC LICENSE v3
`

func main() {
	defaultErrDB.RefreshCache()

	api := getAPI()
	api["release"] = js.FuncOf(func(_ js.Value, _ []js.Value) (_ any) {
		global.Delete("MCLA")
		releaseBgCtx()
		return js.Undefined()
	})
	global.Set("MCLA", api)

	console.Call("log", fmt.Sprintf(LICENSE, version))
	defer fmt.Printf("MCLA-%s unloaded\n", version)

	<-bgCtx.Done()
}

func parseCrashReport(args []js.Value) (report *CrashReport, err error) {
	value := args[0]
	r, err := wrapJsValueAsReader(value)
	if err != nil {
		return
	}
	if report, err = ParseCrashReport(r); err != nil {
		if err == io.EOF {
			// Couldn't find crash report, return null
			return nil, nil
		}
		return
	}
	return
}

func parseLogErrors(args []js.Value) (errs []*JavaError, err error) {
	value := args[0]
	r, err := wrapJsValueAsReader(value)
	if err != nil {
		return
	}
	return ScanJavaErrors(r)
}

func analyzeLogErrors(args []js.Value) (result []*ErrorResult, err error) {
	value := args[0]
	r, err := wrapJsValueAsReader(value)
	if err != nil {
		return
	}
	result = make([]*ErrorResult, 0, 5)
	resCh, ctx := defaultAnalyzer.DoLogStream(bgCtx, r)
	for {
		select {
		case res := <-resCh:
			result = append(result, res)
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		}
	}
	return
}

func analyzeLogErrorsIter(args []js.Value) (iterator js.Value, err error) {
	value := args[0]
	r, err := wrapJsValueAsReader(value)
	if err != nil {
		return
	}
	result, ctx := defaultAnalyzer.DoLogStream(bgCtx, r)
	iterator = NewChannelIteratorContext(ctx, result)
	return
}
