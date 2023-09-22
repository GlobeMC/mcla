
package mcla

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

var (
	ErrUnexpectedIndent = errors.New("Crash report details format incorrect: unexpected indent")
	ErrMissingColon = errors.New("Crash report details format incorrect: missing colon")
)

var (
	crashReportHeader   = strings.ToUpper("---- Minecraft Crash Report ----")
	headThreadKey       = strings.ToUpper("Head")
	affectedLevelKey    = strings.ToUpper("Affected level")
	descriptionHeader   = strings.ToUpper("Description:")
	detailsKeyHeader    = strings.ToUpper("Details:")
	stacktraceHeader    = strings.ToUpper("Stacktrace:")
	threadKeyHeader     = strings.ToUpper("Thread:")
)


func hasIndent(line []byte)(bool){
	return bytes.HasPrefix(line, ([]byte)("\t")) || bytes.HasPrefix(line, ([]byte)("  "))
}

func hasDbIndent(line []byte)(bool){
	return bytes.HasPrefix(line, ([]byte)("\t\t")) || bytes.HasPrefix(line, ([]byte)("    "))
}

// Details:
type ReportDetails map[string][]string

func parseReportDetails(sc *lineScanner)(d ReportDetails, err error){
	if !sc.Scan() {
		return
	}
	return parseReportDetails0(sc)
}

func parseReportDetails0(sc *lineScanner)(d ReportDetails, err error){
	d = make(ReportDetails)
	line := sc.Bytes()
	for {
		if len(line) < 2 || !hasIndent(line) {
			return
		}
		if hasDbIndent(line) {
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
			values = []string{(string)(line)}
		}
		for {
			if !sc.Scan() {
				d.set(key, values)
				return
			}
			if line = sc.Bytes(); len(line) < 2 || !hasIndent(line) {
				d.set(key, values)
				return
			}
			if !hasDbIndent(line) {
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


// -- Head --
type HeadThread struct {
	Thread     string     `json:"thread"`
	Stacktrace Stacktrace `json:"stacktrace"`
}

func parseHeadThread(sc *lineScanner)(res HeadThread, err error){
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

func parseAffectedLevel(sc *lineScanner)(res AffectedLevel, err error){
	if !sc.Scan() {
		return
	}
	firstline := true
	for {
		line := strings.ToUpper(sc.Text())
		switch {
		case strings.HasPrefix(line, "--"):
			return
		case strings.HasPrefix(line, detailsKeyHeader):
			if res.Details, err = parseReportDetails(sc); err != nil {
				return
			}
		case firstline && hasIndent(sc.Bytes()):
			if res.Details, err = parseReportDetails0(sc); err != nil {
				return
			}
		case strings.HasPrefix(line, stacktraceHeader):
			res.Stacktrace = parseStacktrace(sc)
		default:
			if !sc.Scan() {
				return
			}
		}
		firstline = false
	}
	return
}

type DetailsItem struct {
	Details ReportDetails `json:"details"`
}

func parseDetailsItem(sc *lineScanner)(res DetailsItem, err error){
	if !sc.Scan() {
		return
	}
	firstline := true
	for {
		line := strings.ToUpper(sc.Text())
		switch {
		case strings.HasPrefix(line, "--"):
			return
		case strings.HasPrefix(line, detailsKeyHeader):
			if res.Details, err = parseReportDetails(sc); err != nil {
				return
			}
		case firstline && hasIndent(sc.Bytes()):
			if res.Details, err = parseReportDetails0(sc); err != nil {
				return
			}
		default:
			if !sc.Scan() {
				return
			}
		}
		firstline = false
	}
	return
}

type CrashReport struct {     // ---- Minecraft Crash Report ----
	Description   string        `json:"description"`     // Description:
	Error         *JavaError    `json:"error"`
	HeadThread    HeadThread    `json:"head"`            // -- Head --
	AffectedLevel AffectedLevel `json:"affectedLevel"`   // -- Affected level --
	OtherDetails  map[string]DetailsItem `json:"others"` // -- <KEY> --
}

func ParseCrashReport(r io.Reader)(report *CrashReport, err error){
	sc := newLineScanner(r)
	for {
		if !sc.Scan() {
			if err = sc.Err(); err == nil {
				err = io.EOF
			}
			return
		}
		if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(sc.Text())), crashReportHeader) {
			break
		}
	}
	report = &CrashReport{
		OtherDetails: make(map[string]DetailsItem),
	}
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
			report.Error = parseJavaError(sc)
		case strings.HasPrefix(uline, descriptionHeader):
			if flag != 0 {
				return nil, errors.New("Key `Description` duplicated")
			}
			report.Description = strings.TrimSpace(line[len(descriptionHeader):])
			if !sc.Scan() {
				return
			}
			flag = 1
		case strings.HasPrefix(uline, "-- ") && strings.HasSuffix(uline, " --"):
			name := strings.ToUpper((string)(uline[len("-- "):len(uline) - len(" --")]))
			switch {
			case name == headThreadKey:
				if report.HeadThread, err = parseHeadThread(sc); err != nil {
					return
				}
			case name == affectedLevelKey:
				if report.AffectedLevel, err = parseAffectedLevel(sc); err != nil {
					return
				}
				report.OtherDetails[affectedLevelKey] = DetailsItem{report.AffectedLevel.Details}
			default:
				var details DetailsItem
				if details, err = parseDetailsItem(sc); err != nil {
					return
				}
				report.OtherDetails[name] = details
			}
		default:
			if !sc.Scan() {
				return
			}
		}
	}
	return
}

func (report *CrashReport)GetDetails(key string)(value DetailsItem){
	return report.OtherDetails[strings.ToUpper(key)]
}
