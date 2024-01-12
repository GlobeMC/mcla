package mcla

import (
	"context"
	"errors"
	"io"
	"sync"
)

type SolutionPossibility struct {
	ErrorDesc *ErrorDesc `json:"errorDesc"`
	Match     float32    `json:"match"`
}

type ErrorResult struct {
	Error   *JavaError            `json:"error"`
	Matched []SolutionPossibility `json:"matched"`
	File    string                `json:"file,omitempty"`
}

var (
	ErrCrashReportIncomplete = errors.New("Crashreport is incomplete")
)

type Analyzer struct {
	DB ErrorDB
}

func NewAnalyzer(db ErrorDB) (a *Analyzer) {
	return &Analyzer{
		DB: db,
	}
}

func (a *Analyzer) DoError(jerr *JavaError) (matched []SolutionPossibility, err error) {
	for jerr != nil {
		epkg, ecls := rsplit(jerr.Class, '.')
		a.DB.ForEachErrors(func(e *ErrorDesc) (err error) {
			sol := SolutionPossibility{
				ErrorDesc: e,
			}
			epkg2, ecls2 := rsplit(e.Error, '.')
			ignoreErrorTyp := len(ecls2) == 0 || ecls2 == "*"
			if !ignoreErrorTyp && ecls2 == ecls { // error type weight: 10%
				if epkg2 == "*" || epkg == epkg2 {
					sol.Match = 0.1 // 10%
				} else {
					sol.Match = 0.05 // 5%
				}
			}
			if len(e.Message) == 0 { // when ignore error message, error type provide 100% score weight
				sol.Match /= 10 / 100
			} else {
				jemsg, _ := split(jerr.Message, '\n')
				matches := lineMatchPercent(jemsg, e.Message) // error message weight: 90%
				if ignoreErrorTyp {
					sol.Match = matches // or when ignore error type, it provide 100% score weight
				} else {
					sol.Match += matches * 0.9
				}
			}
			if sol.Match != 0 { // have any matches
				matched = append(matched, sol)
			}
			return
		})
		jerr = jerr.CausedBy
	}
	if matched == nil {
		matched = make([]SolutionPossibility, 0)
	}
	return
}

func (a *Analyzer) DoLogStream(c context.Context, r io.Reader) (<-chan *ErrorResult, context.Context) {
	result := make(chan *ErrorResult, 3)
	ctx, cancel := context.WithCancelCause(c)
	go func() {
		defer close(result)
		var wg sync.WaitGroup
		resCh, errCh := ScanJavaErrorsIntoChan(r)
	LOOP:
		for {
			select {
			case jerr := <-resCh:
				if jerr == nil {
					break LOOP
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					res := &ErrorResult{
						Error: jerr,
					}
					var err error
					if res.Matched, err = a.DoError(jerr); err != nil {
						cancel(err)
						return
					}
					select {
					case result <- res:
					case <-ctx.Done():
					}
				}()
			case err := <-errCh:
				cancel(err)
				return
			case <-ctx.Done():
				return
			}
		}
		wg.Wait()
	}()
	return result, ctx
}
