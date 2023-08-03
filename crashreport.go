
package mcla

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

var (
	ErrUnexpectedIndent = errors.New("Crash report details format incorrect: unexpected indent")
	ErrMissingColon = errors.New("Crash report details format incorrect: missing colon")
)

var (
	crashReportHeader   = strings.ToUpper("---- Minecraft Crash Report ----")
	headThreadHeader    = strings.ToUpper("-- Head --")
	affectedLevelHeader = strings.ToUpper("-- Affected level --")
	lastReloadHeader    = strings.ToUpper("-- Last reload --")
	systemDetailsHeader = strings.ToUpper("-- System Details --")
	descriptionHeader   = strings.ToUpper("Description:")
	detailsKeyHeader    = strings.ToUpper("Details:")
	stacktraceHeader    = strings.ToUpper("Stacktrace:")
	threadKeyHeader     = strings.ToUpper("Thread:")
)

var (
	stackInfoMatcher = regexp.MustCompile(`^\s*at\s+([\w\d$_]+(?:\.[\w\d$_]+)*)\.([\w\d$_<>]+)`)
)

// Details:
type ReportDetails map[string][]string

func parseReportDetails(sc *bufio.Scanner)(d ReportDetails, err error){
	d = make(ReportDetails)
	if !sc.Scan() {
		return
	}
	line := sc.Bytes()
	for {
		if len(line) < 2 || line[0] != '\t' {
			return
		}
		if line[1] == '\t' {
			return nil, ErrUnexpectedIndent
		}
		line = line[1:]
		i := bytes.IndexByte(line, ':')
		if i < 0 {
			return nil, ErrMissingColon
		}
		var (
			key string = (string)(bytes.TrimSpace(line[:i]))
			values []string
		)
		if line = bytes.TrimSpace(line[i + 1:]); len(line) > 0 {
			values = strings.Split((string)(line), ":")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
		}
		for {
			if !sc.Scan() {
				d.set(key, values)
				return
			}
			if line = sc.Bytes(); len(line) < 2 || line[0] != '\t' {
				d.set(key, values)
				return
			}
			if line[1] != '\t' {
				d.set(key, values)
				break
			}
			values = append(values, (string)(bytes.TrimSpace(line)))
		}
	}
}

func (d ReportDetails)set(key string, values []string){
	d[strings.ToUpper(key)] = values
}

func (d ReportDetails)Has(key string)(ok bool){
	_, ok = d[strings.ToUpper(key)]
	return
}

func (d ReportDetails)Get(key string)(value string){
	values, ok := d[strings.ToUpper(key)]
	if !ok {
		return ""
	}
	return strings.Join(values, "\n")
}

func (d ReportDetails)GetValues(key string)(values []string){
	return d[strings.ToUpper(key)]
}


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

// -- Head --
type HeadThread struct {
	Thread     string     `json:"thread"`
	Stacktrace Stacktrace `json:"stacktrace"`
}

func parseHeadThread(sc *bufio.Scanner)(res HeadThread, err error){
	if !sc.Scan() {
		return
	}
	for {
		oline := sc.Text()
		line := strings.ToUpper(oline)
		switch {
		case strings.HasPrefix(line, "--"):
			return
		case strings.HasPrefix(line, threadKeyHeader):
			res.Thread = strings.TrimSpace(oline[len(threadKeyHeader):])
			if !sc.Scan() {
				return
			}
		case strings.HasPrefix(line, stacktraceHeader):
			res.Stacktrace = parseStacktrace(sc)
		default:
			if !sc.Scan() {
				return
			}
		}
	}
	return
}

// -- Affected level --
type AffectedLevel struct {
	Details    ReportDetails `json:"details"`
	Stacktrace Stacktrace    `json:"stacktrace"`
}

func parseAffectedLevel(sc *bufio.Scanner)(res AffectedLevel, err error){
	if !sc.Scan() {
		return
	}
	for {
		line := strings.ToUpper(sc.Text())
		switch {
		case strings.HasPrefix(line, "--"):
			return
		case strings.HasPrefix(line, detailsKeyHeader):
			if res.Details, err = parseReportDetails(sc); err != nil {
				return
			}
		case strings.HasPrefix(line, stacktraceHeader):
			res.Stacktrace = parseStacktrace(sc)
		default:
			if !sc.Scan() {
				return
			}
		}
	}
	return
}

type LastReload struct {
	Details ReportDetails `json:"details"`
}

func parseLastReload(sc *bufio.Scanner)(res LastReload, err error){
	if !sc.Scan() {
		return
	}
	for {
		line := strings.ToUpper(sc.Text())
		switch {
		case strings.HasPrefix(line, "--"):
			return
		case strings.HasPrefix(line, detailsKeyHeader):
			if res.Details, err = parseReportDetails(sc); err != nil {
				return
			}
		default:
			if !sc.Scan() {
				return
			}
		}
	}
	return
}

type SystemDetails struct {
	Details ReportDetails `json:"details"`
}

func parseSystemDetails(sc *bufio.Scanner)(res SystemDetails, err error){
	if !sc.Scan() {
		return
	}
	for {
		line := strings.ToUpper(sc.Text())
		switch {
		case strings.HasPrefix(line, "--"):
			return
		case strings.HasPrefix(line, detailsKeyHeader):
			if res.Details, err = parseReportDetails(sc); err != nil {
				return
			}
		default:
			if !sc.Scan() {
				return
			}
		}
	}
	return
}

type CrashReport struct {     // ---- Minecraft Crash Report ----
	Description   string        `json:"description"`    // Description:
	ErrorClass    string        `json:"error_class"`
	ErrorMessage  string        `json:"error_message"`
	ErrStacktrace Stacktrace    `json:"error_stacktrace"`
	HeadThread    HeadThread    `json:"head"`           // -- Head --
	AffectedLevel AffectedLevel `json:"affected_level"` // -- Affected level --
	LastReload    LastReload    `json:"last_reload"`    // -- Last reload --
	SystemDetails SystemDetails `json:"system_details"` // -- System Details --
}

func ParseCrashReport(r io.Reader)(report *CrashReport, err error){
	sc := bufio.NewScanner(r) // default is scan lines
	for {
		if !sc.Scan() {
			return nil, sc.Err()
		}
		if strings.HasPrefix(strings.ToUpper(sc.Text()), crashReportHeader) {
			break
		}
	}
	report = new(CrashReport)
	if !sc.Scan() {
		return
	}
	var line string
	var flag int = 0
	for {
		line = sc.Text()
		uline := strings.ToUpper(line)
		switch {
		case len(uline) == 0 && flag == 1:
			flag = 2
			if !sc.Scan() {
				return
			}
			line = sc.Text()
			if !sc.Scan() {
				return
			}
			i := strings.IndexByte(line, ':')
			if i == -1 {
				report.ErrorMessage = line
			}else{
				report.ErrorClass, report.ErrorMessage = line[:i], strings.TrimSpace(line[i + 1:])
			}
			report.ErrStacktrace = parseStacktrace(sc)
		case strings.HasPrefix(uline, descriptionHeader):
			if flag != 0 {
				return nil, errors.New("Key `Description` duplicated")
			}
			report.Description = strings.TrimSpace(line[len(descriptionHeader):])
			if !sc.Scan() {
				return
			}
			flag = 1
		case strings.HasPrefix(uline, headThreadHeader):
			if report.HeadThread, err = parseHeadThread(sc); err != nil {
				return
			}
		case strings.HasPrefix(uline, affectedLevelHeader):
			if report.AffectedLevel, err = parseAffectedLevel(sc); err != nil {
				return
			}
		case strings.HasPrefix(uline, lastReloadHeader):
			if report.LastReload, err = parseLastReload(sc); err != nil {
				return
			}
		case strings.HasPrefix(uline, systemDetailsHeader):
			if report.SystemDetails, err = parseSystemDetails(sc); err != nil {
				return
			}
		default:
			if !sc.Scan() {
				return
			}
		}
	}
	return
}
