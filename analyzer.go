
package mcla

import (
	"errors"
)

type SolutionPossibility struct {
	Solutions []int   `json:"solutions"`
	Match     float32 `json:"match"`
}

var (
	ErrCrashReportIncomplete = errors.New("Crashreport is incomplete")
)

type Analyzer struct {
	DB ErrorDB
}

func NewAnalyzer(db ErrorDB)(a *Analyzer){
	return &Analyzer{
		DB: db,
	}
}

func (a *Analyzer)DoError(jerr *JavaError)(sols []SolutionPossibility, err error){
	for jerr != nil {
		epkg, ecls := rsplit(jerr.Class, '.')
		a.DB.ForEachErrors(func(e *ErrorDesc)(err error){
			sol := SolutionPossibility{
				Solutions: e.Solutions,
			}
			epkg2, ecls2 := rsplit(e.Error, '.')
			ignoreErrorTyp := len(ecls2) == 0 || ecls2 == "*"
			if !ignoreErrorTyp && ecls2 == ecls { // error type weight: 10%
				if epkg2 == "*" || epkg == epkg2 {
					sol.Match = 0.1 // 10%
				}else{
					sol.Match = 0.05 // 5%
				}
			}
			if len(e.Message) == 0 { // when ignore error message, error type provide 100% score weight
				sol.Match /= 10 / 100
			}else{
				jemsg, _ := split(jerr.Message, '\n')
				matches := lineMatchPercent(jemsg, e.Message) // error message weight: 90%
				if ignoreErrorTyp {
					sol.Match = matches // or when ignore error type, it provide 100% score weight
				}else{
					sol.Match += matches * 0.9
				}
			}
			if sol.Match >= 0.5 { // at least have 50% matches
				sols = append(sols, sol)
			}
			return
		})
		jerr = jerr.CausedBy
	}
	if sols == nil {
		sols = make([]SolutionPossibility, 0)
	}
	return
}
