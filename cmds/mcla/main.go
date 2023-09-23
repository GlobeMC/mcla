
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kmcsr/mcla"
)

func printf(format string, args ...any){
	if format[len(format) - 1] != '\n' {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func main(){
	if len(os.Args) <= 1 {
		help()
		return
	}
	subcmd := os.Args[1]
	switch subcmd {
	case "parseCrashReport":
		if len(os.Args) <= 2 {
			printf("[ERROR]: Must give the crashreport's filename as the second argument")
			os.Exit(1)
		}
		filename := os.Args[2]
		fd, err := os.Open(filename)
		if err != nil {
			printf("Error when opening report file: %v", err)
			os.Exit(1)
		}
		report, err := mcla.ParseCrashReport(fd)
		fd.Close()
		if err != nil {
			printf("Error when parsing report file: %v", err)
			os.Exit(1)
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(report); err != nil {
			printf("\nError when encoding report file as json: %v", err)
			os.Exit(1)
		}
	case "analyzeErrors":
		if len(os.Args) <= 2 {
			return
		}
		files := os.Args[2:]
		for _, name := range files {
			analysisAndOutput(name)
		}
	case "help":
		help()
	default:
		printf("[ERROR]: Unknown command %q", subcmd)
		help()
		os.Exit(1)
	}
}

func analysisAndOutput(file string){
	fd, err := os.Open(file)
	if err != nil {
		printf("Error when opening file %q: %v", file, err)
		os.Exit(1)
	}
	defer fd.Close()
	resCh, errCh := analyzeLogErrors(fd)
	if err != nil {
		printf("Error when analyzing file %q: %v", file, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	ok := false
LOOP_RES:
	for {
		select {
		case res := <-resCh:
			if res == nil { // done
				break LOOP_RES
			}
			ok = true
			res.File = file
			if err = encoder.Encode(res); err != nil {
				printf("\nError when encoding report file as json: %v", err)
				os.Exit(1)
			}
		case err := <-errCh:
			printf("Error when analyzing file %q: %v", file, err)
			os.Exit(1)
		}
	}
	if !ok {
		printf("No any error was found")
		os.Exit(1)
	}
}
