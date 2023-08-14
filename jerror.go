
package mcla

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	javaErrorMatcher = regexp.MustCompile(`^([\w\d$_]+(?:\.[\w\d$_]+)*):\s+(.*)$`)
	stackInfoMatcher = regexp.MustCompile(`^\s+at\s+([\w\d$_]+(?:\.[\w\d$_]+)*)\.([\w\d$_<>]+)`)
)

type (
	StackInfo struct {
		Raw     string `json:"raw"`
		Class   string `json:"class"`
		Method  string `json:"method"`
	}

	// Stacktrace:
	Stacktrace []StackInfo
)

func parseStackInfoFrom(line string)(s StackInfo, ok bool){
	res := stackInfoMatcher.FindStringSubmatch(line)
	if res == nil {
		return
	}
	s.Raw = line
	s.Class = res[1]
	s.Method = res[2]
	ok = true
	return
}

func parseStacktrace(sc *bufio.Scanner)(st Stacktrace){
	var (
		info StackInfo
		ok bool
	)
	for sc.Scan() {
		line := sc.Text()
		if info, ok = parseStackInfoFrom(line); !ok {
			return
		}
		st = append(st, info)
	}
	return
}

type JavaError struct {
	Class      string     `json:"class"`
	Message    string     `json:"message"`
	Stacktrace Stacktrace `json:"stacktrace"`
	CausedBy   *JavaError `json:"caused_by"`
}

func parseJavaError(sc *bufio.Scanner)(je *JavaError){
	if !sc.Scan() {
		return
	}
	return parseJavaError0(sc.Text(), sc)
}

func parseJavaError0(line string, sc *bufio.Scanner)(je *JavaError){
	je = new(JavaError)
	i := strings.IndexByte(line, ':')
	if i == -1 {
		je.Message = line
	}else{
		je.Class, je.Message = line[:i], strings.TrimSpace(line[i + 1:])
	}
	je.Stacktrace = parseStacktrace(sc)
	line = sc.Text()
	if line, ok := strings.CutPrefix(line, "Caused by: "); ok {
		je.CausedBy = parseJavaError0(line, sc)
	}
	return
}

func ScanJavaErrors(r io.Reader)(errs []*JavaError){
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		return
	}
	var line string
	for {
		line = sc.Text()
		emsg := javaErrorMatcher.FindStringSubmatch(line)
		if emsg == nil {
			if !sc.Scan() {
				break
			}
			continue
		}
		st := parseStacktrace(sc)
		if st != nil { // if stacktrace exists 
			je := &JavaError{
				Class: emsg[1],
				Message: emsg[2],
				Stacktrace: st,
			}
			if line, ok := strings.CutPrefix(sc.Text(), "Caused by: "); ok {
				je.CausedBy = parseJavaError0(line, sc)
			}
			errs = append(errs, je)
		}
	}
	if errs == nil { // to ensure the json encoding is an array
		errs = make([]*JavaError, 0)
	}
	return
}
