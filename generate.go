package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func generateCmd() *cobra.Command {
	var file string
	var number int
	var seed int64
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate random tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var w io.Writer = os.Stdout
			if file != "" {
				f, err := os.Create(file)
				if err != nil {
					return fmt.Errorf("could not create tokens file: %w", err)
				}
				defer f.Close()
				w = f
			}
			return generate(w, number, seed)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "file to save the generated tokens. will write to stdout if omitted")
	cmd.Flags().IntVarP(&number, "number", "n", 1e7, "number of tokens to generate")
	cmd.Flags().Int64VarP(&seed, "seed", "s", -1, "seed for random token generation. passing a negative value will use the current time as a seed")
	return cmd
}

func generate(w io.Writer, number int, seed int64) error {
	if seed < 0 {
		seed = time.Now().Unix()
	}
	rand.Seed(seed)

	bw := bufio.NewWriter(w)
	for i := 0; i < number; i++ {
		_, err := bw.WriteString(randomToken() + "\n")
		if err != nil {
			return fmt.Errorf("could not write token: %w", err)
		}
	}
	err := bw.Flush()
	if err != nil {
		return fmt.Errorf("could not flush writer: %w", err)
	}

	return nil
}

func randomToken() string {
	var sb strings.Builder
	for i := 0; i < 7; i++ {
		sb.WriteRune('a' + rune(rand.Intn(26)))
	}
	return sb.String()
}
