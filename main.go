package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "CLI for token management",
	}
	cmd.AddCommand(generateCmd())
	cmd.AddCommand(importCmd())
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		os.Exit(1)
	}
}
