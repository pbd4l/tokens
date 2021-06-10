package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Generate generates random tokens and writes them to a file
func Generate(args []string) error {
	flagSet := flag.NewFlagSet("generate", flag.ContinueOnError)

	file := flagSet.String("file", "tokens.txt", "file to write the tokens.")
	n := flagSet.Int("n", 1e7, "number of tokens to generate.")
	seed := flagSet.Int64("seed", -1, "seed for random token generation. passing a negative value will use the current time as a seed.")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("could not parse flags: %w", err)
	}

	f, err := os.Create(*file)
	if err != nil {
		return fmt.Errorf("could not create tokens file: %w", err)
	}
	defer f.Close()

	return generate(f, *n, *seed)
}

func generate(w io.Writer, n int, seed int64) error {
	if seed < 0 {
		seed = time.Now().Unix()
	}
	rand.Seed(seed)

	bw := bufio.NewWriter(w)
	for i := 0; i < n; i++ {
		token, err := randomToken()
		if err != nil {
			return fmt.Errorf("could not generate token: %w", err)
		}
		if _, err := bw.WriteString(token + "\n"); err != nil {
			return fmt.Errorf("could not write token: %w", err)
		}
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("could not flush writer: %w", err)
	}

	return nil
}

func randomToken() (string, error) {
	var sb strings.Builder
	for i := 0; i < 7; i++ {
		if _, err := sb.WriteRune('a' + rune(rand.Intn(26))); err != nil {
			return "", fmt.Errorf("could not write token rune: %w", err)
		}
	}
	return sb.String(), nil
}
