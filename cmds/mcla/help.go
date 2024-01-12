package main

import (
	"fmt"
)

const LICENSE = `mcla (Minecraft Log Analyzer) v%s
Copyright (C) 2023 <zyxkad@gmail.com> all rights reserved
Under GNU GENERAL PUBLIC LICENSE v3
`

const HELP_MESSAGE = `
Usage:
   mcla <subcommand> [<subcmd args>...]

Subcommands:
   - parseCrashReport <filename>
   - analyzeErrors [<filename>...]
`

func help() {
	fmt.Printf(LICENSE, version)
	fmt.Print(HELP_MESSAGE)
}
