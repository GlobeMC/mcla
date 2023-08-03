
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kmcsr/mcla"
)

func main(){
	if len(os.Args) <= 1 {
		help()
		return
	}
	subcmd := os.Args[1]
	switch subcmd {
	case "parseCrashReport":
		if len(os.Args) <= 2 {
			fmt.Printf("[ERROR]: Must give the crashreport's filename as the second argument")
			os.Exit(1)
		}
		filename := os.Args[2]
		fd, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error when opening report file:", err)
			os.Exit(1)
		}
		report, err := mcla.ParseCrashReport(fd)
		fd.Close()
		if err != nil {
			fmt.Println("Error when parsing report file:", err)
			os.Exit(1)
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(report); err != nil {
			fmt.Println("\nError when encoding report file as json:", err)
			os.Exit(1)
		}
	case "help":
		help()
	default:
		fmt.Printf("[ERROR]: Unknown command %q\n", subcmd)
		help()
		os.Exit(1)
	}
}
