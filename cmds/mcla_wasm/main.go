
//go:build tinygo.wasm
package main

import (
	"context"
	"fmt"
	"io"
	"sync"
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
		"analyzeLogErrorsIter": asyncFuncOf(func(_ js.Value, args []js.Value)(res any){
			return analyzeLogErrorsIter(args)
		}),
		"setGhDbPrefix": js.FuncOf(func(_ js.Value, args []js.Value)(res any){
			prefix := args[0]
			prefixStr := prefix.String()
			fmt.Printf("Set database as %q\n", prefixStr)
			defaultErrDB.Prefix = prefixStr
			return
		}),
	}
}

func main(){
	api := getAPI()
	api["release"] = js.FuncOf(func(_ js.Value, _ []js.Value)(_ any){
		global.Delete("MCLA")
		releaseBgCtx()
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

type ErrorResult struct {
	Error   *JavaError            `json:"error"`
	Matched []SolutionPossibility `json:"matched"`
}

func analyzeLogErrors(args []js.Value)(result []ErrorResult){
	value := args[0]
	r := wrapJsValueAsReader(value)
	errs := ScanJavaErrors(r)
	result = make([]ErrorResult, len(errs))

	ctx, cancel := context.WithCancelCause(bgCtx)
	doneCh := make(chan struct{}, len(errs))
	for i, jerr := range errs {
		go func(){
			defer func(){
				doneCh <- struct{}{}
			}()
			var (
				res ErrorResult
				err error
			)
			res.Error = jerr
			if res.Matched, err = defaultAnalyzer.DoError(jerr); err != nil {
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

func analyzeLogErrorsIter(args []js.Value)(iterator js.Value){
	value := args[0]
	r := wrapJsValueAsReader(value)
	result := make(chan ErrorResult, 3)
	ctx, cancel := context.WithCancelCause(bgCtx)
	iterator = NewChannelIteratorContext(ctx, result)
	go func(){
		defer close(result)
		var wg sync.WaitGroup
		errCh := ScanJavaErrorsIntoChan(r)
		for jerr := range errCh {
			wg.Add(1)
			go func(){
				defer wg.Done()
				var (
					res ErrorResult
					err error
				)
				res.Error = jerr
				if res.Matched, err = defaultAnalyzer.DoError(jerr); err != nil {
					cancel(err)
					return
				}
				select {
				case result <- res:
				case <-bgCtx.Done():
				}
			}()
		}
		wg.Wait()
	}()
	return
}
