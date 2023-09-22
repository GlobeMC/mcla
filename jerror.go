
package mcla

import (
	"io"
	"regexp"
	"strings"
)

var (
	javaErrorMatcher = regexp.MustCompile(`^([\w\d$_]+(?:\.[\w\d$_]+)*):\s+(.*)$`)
	stackInfoMatcher = regexp.MustCompile(`^\s+at\s+([\w\d$_]+(?:\.[\w\d$_]+)*)\.([\w\d$_<>]+)`)
)

type (
	JavaError struct {
		Class      string     `json:"class"`
		Message    string     `json:"message"`
		Stacktrace Stacktrace `json:"stacktrace"`
		CausedBy   *JavaError `json:"causedBy"`

		// extra infos
	  LineNo int `json:"lineNo"` // which line did the error start
	}

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

func parseStacktrace(sc *lineScanner)(st Stacktrace){
	if !sc.Scan() {
		return
	}
	return parseStacktrace0(sc)
}

func parseStacktrace0(sc *lineScanner)(st Stacktrace){
	var (
		info StackInfo
		ok bool
	)
	for {
		line := sc.Text()
		if info, ok = parseStackInfoFrom(line); !ok {
			return
		}
		st = append(st, info)
		if !sc.Scan() {
			return
		}
	}
	return
}


func parseJavaError(sc *lineScanner)(je *JavaError){
	if !sc.Scan() {
		return
	}
	return parseJavaError0(sc.Text(), sc)
}

func parseJavaError0(line string, sc *lineScanner)(je *JavaError){
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

func scanJavaErrors(r io.Reader, cb func(*JavaError)){
	sc := newLineScanner(r)
	if !sc.Scan() {
		return
	}
	var (
		line string
		lineNo int
	)
	for {
		line = sc.Text()
		lineNo = sc.Count()
		emsg := javaErrorMatcher.FindStringSubmatch(line)
		if !sc.Scan() {
			break
		}
		if emsg == nil {
			continue
		}
		for {
			l2 := sc.Text()
			if stackInfoMatcher.MatchString(l2) {
				break
			}
			if em := javaErrorMatcher.FindStringSubmatch(l2); em != nil {
				line = l2
				emsg = em
			}else{
				emsg[2] += "\n" + l2
			}
			if !sc.Scan() {
				break
			}
		}
		st := parseStacktrace0(sc)
		if st != nil { // if stacktrace exists 
			je := &JavaError{
				Class: emsg[1],
				Message: emsg[2],
				Stacktrace: st,
				LineNo: lineNo,
			}
			if line, ok := strings.CutPrefix(sc.Text(), "Caused by: "); ok {
				je.CausedBy = parseJavaError0(line, sc)
			}
			cb(je)
		}
	}
	return
}

func ScanJavaErrors(r io.Reader)(errs []*JavaError){
	errs = make([]*JavaError, 0, 3)
	scanJavaErrors(r, func(je *JavaError){
		errs = append(errs, je)
	})
	return
}

func ScanJavaErrorsIntoChan(r io.Reader)(errs <-chan *JavaError){
	errs0 := make(chan *JavaError, 3)
	go func(){
		defer close(errs0)
		scanJavaErrors(r, func(je *JavaError){
			errs0 <- je
		})
	}()
	return errs0
}
