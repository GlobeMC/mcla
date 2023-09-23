
package main

import (
	"io"
	"sync"

	"github.com/kmcsr/mcla"
)

var analyzer mcla.Analyzer

type ErrorResult struct {
	Error   *mcla.JavaError            `json:"error"`
	Matched []mcla.SolutionPossibility `json:"matched"`
	File    string                     `json:"file"`
}

func analyzeLogErrors(r io.Reader)(<-chan *ErrorResult, <-chan error){
	resCh := make(chan *ErrorResult, 3)
	errCh := make(chan error, 1)
	go func(){
		defer close(resCh)
		var wg sync.WaitGroup
		jerrCh, errC := mcla.ScanJavaErrorsIntoChan(r)
	LOOP:
		for {
			select {
			case err := <-errC:
				errCh <- err
				return
			case jerr := <-jerrCh:
				if jerr == nil {
					break LOOP
				}
				wg.Add(1)
				go func(){
					defer wg.Done()
					var err error
					res := &ErrorResult{
						Error: jerr,
					}
					if res.Matched, err = analyzer.DoError(jerr); err != nil {
						errCh <- err
						return
					}
					resCh <- res
				}()
			}
		}
		wg.Wait()
	}()
	return resCh, errCh
}
