
package main

import (
	"fmt"
)

const LICENSE = `mcla (Minecraft Log Analyzer) Copyright (C) 2023 <zyxkad@gmail.com> all rights reserved
Under GNU GENERAL PUBLIC LICENSE v3
`

const HELP_MESSAGE = `
Usage:
   mcla <subcommand> [<subcommand_args>...]

Subcommands:
   - parseCrashReport <filename>
`

func help(){
	fmt.Print(LICENSE)
	fmt.Print(HELP_MESSAGE)
}
