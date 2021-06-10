package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
)

// Import imports tokens from a file into a postgres database
func Import(args []string) error {
	flagSet := flag.NewFlagSet("import", flag.ContinueOnError)

	dsn := flagSet.String("dsn", "", "postgres data-source-name.")
	file := flagSet.String("file", "tokens.txt", "file to import the tokens from.")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("could not parse flags: %w", err)
	}
	if *dsn == "" {
		return fmt.Errorf("flag -dsn is required")
	}

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS tokens (token CHAR(7) PRIMARY KEY)"); err != nil {
		return fmt.Errorf("could not create/ensure table: %w", err)
	}

	tokensFile, err := os.Open(*file)
	if err != nil {
		return fmt.Errorf("could not open tokens file: %w", err)
	}
	defer tokensFile.Close()

	err = importTokens(db, tokensFile)
	if err != nil {
		return fmt.Errorf("could not import tokens: %w", err)
	}

	return nil
}

var tokenRe = regexp.MustCompile(`^[a-z]{7}$`)

func importTokens(db *sql.DB, r io.Reader) error {
	bs := 10000

	scanner := bufio.NewScanner(r)

	var sb strings.Builder
	sb.WriteString("INSERT INTO tokens (token) VALUES ")
	i := 0
	args := make([]interface{}, 0, bs)
	for scanner.Scan() {
		token := scanner.Text()
		if !tokenRe.MatchString(token) {
			log.Printf("token %s is invalid, skipping\n", token)
			continue
		}

		if i == 0 {
			sb.WriteString(fmt.Sprintf("($%d)", i+1))
		} else {
			sb.WriteString(fmt.Sprintf(",($%d)", i+1))
		}
		i++
		args = append(args, token)

		if i == bs {
			sb.WriteString(" ON CONFLICT DO NOTHING")
			_, err := db.Exec(sb.String(), args...)
			if err != nil {
				return fmt.Errorf("could not insert tokens: %w", err)
			}
			sb.Reset()
			sb.WriteString("INSERT INTO tokens (token) VALUES ")
			i = 0
			args = make([]interface{}, 0, bs)
		}
	}

	if i > 0 {
		sb.WriteString(" ON CONFLICT DO NOTHING")
		_, err := db.Exec(sb.String(), args[:i]...)
		if err != nil {
			return fmt.Errorf("could not insert tokens: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("could not scan tokens: %w", err)
	}

	return nil
}
