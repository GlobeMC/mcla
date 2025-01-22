package mcla

import (
	"io"
	"regexp"
	"strings"
)

var (
	javaErrorMatcher = regexp.MustCompile(`^\s*(?:Exception in thread "[^"]+"\s+)?([\w\d$_]+(?:\.[\w\d$_]+)+)(?::\s+(.*))?$`)
	stackInfoMatcher = regexp.MustCompile(`^\s*at\s+(?:.+/)?([\w\d$_]+(?:\.[\w\d$_]+)+)\.([\w\d$_<>]+)(?:\s*\((.+)\))?`)
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
		Raw    string `json:"raw"`
		Class  string `json:"class"`
		Method string `json:"method"`
	}

	// Stacktrace:
	Stacktrace []StackInfo
)

func parseStackInfoFrom(line string) (s StackInfo, ok bool) {
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

func parseStacktrace(sc *lineScanner) (st Stacktrace) {
	if !sc.Scan() {
		return
	}
	return parseStacktrace0(sc)
}

func parseStacktrace0(sc *lineScanner) (st Stacktrace) {
	var (
		info StackInfo
		ok   bool
	)
	for {
		line := sc.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "... ") && strings.HasSuffix(line, " more") {
			sc.Scan() // move to the next line
			return
		}
		if info, ok = parseStackInfoFrom(line); !ok {
			return
		}
		st = append(st, info)
		if !sc.Scan() {
			return
		}
	}
}

func parseJavaError(sc *lineScanner) (je *JavaError) {
	if !sc.Scan() {
		return
	}
	return parseJavaError0(sc.Text(), sc)
}

func parseJavaError0(line string, sc *lineScanner) (je *JavaError) {
	je = new(JavaError)
	i := strings.IndexByte(line, ':')
	if i == -1 {
		je.Class = line
	} else {
		je.Class, je.Message = line[:i], strings.TrimSpace(line[i+1:])
	}
	je.LineNo = sc.Count()
	je.Stacktrace = parseStacktrace(sc)
	if line, ok := strings.CutPrefix(strings.TrimSpace(sc.Text()), "Caused by: "); ok {
		je.CausedBy = parseJavaError0(line, sc)
	}
	return
}

func scanJavaErrors(r io.Reader, cb func(*JavaError)) (err error) {
	sc := newLineScanner(r)
	if !sc.Scan() {
		return sc.Err()
	}
	var (
		line   string
		lineNo int
	)
	for {
		line = sc.Text()
		lineNo = sc.Count()
		emsg := javaErrorMatcher.FindStringSubmatch(line)
		if !sc.Scan() {
			return sc.Err()
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
			} else {
				emsg[2] += "\n" + l2
			}
			if !sc.Scan() {
				break
			}
		}
		st := parseStacktrace0(sc)
		if st != nil { // if stacktrace exists
			je := &JavaError{
				Class:      emsg[1],
				Message:    emsg[2],
				Stacktrace: st,
				LineNo:     lineNo,
			}
			if line, ok := strings.CutPrefix(strings.TrimSpace(sc.Text()), "Caused by: "); ok {
				je.CausedBy = parseJavaError0(line, sc)
			}
			cb(je)
		}
	}
}

func ScanJavaErrors(r io.Reader) (res []*JavaError, err error) {
	res = make([]*JavaError, 0, 3)
	err = scanJavaErrors(r, func(je *JavaError) {
		res = append(res, je)
	})
	return
}

func ScanJavaErrorsIntoChan(r io.Reader) (<-chan *JavaError, <-chan error) {
	resCh := make(chan *JavaError, 3)
	errCh := make(chan error, 0)
	go func() {
		defer close(resCh)
		err := scanJavaErrors(r, func(je *JavaError) {
			resCh <- je
		})
		if err != nil {
			errCh <- err
		}
	}()
	return resCh, errCh
}
