package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		exitUnrecognisedCmd()
	}

	var cmd func([]string) error
	switch os.Args[1] {
	case "generate":
		cmd = Generate
	case "import":
		cmd = Import
	default:
		exitUnrecognisedCmd()
	}

	if err := cmd(os.Args[2:]); err != nil {
		log.Fatal(err)
	}
}

func exitUnrecognisedCmd() {
	fmt.Fprintf(os.Stderr, `
Usage: tokens <command>

Commands:
  generate    Generate random tokens and write them to a file
  import      Import tokens from a file into a database

Run 'tokens <command> -help' for more information about a command.

`)
	os.Exit(1)
}
