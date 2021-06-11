package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {
	var dsn string
	var file string
	var bs int
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import tokens to a postgres database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var r io.Reader = os.Stdin
			if file != "" {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("could not open tokens file: %w", err)
				}
				defer f.Close()
				r = f
			}
			return importTokens(r, dsn, bs, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVarP(&dsn, "dsn", "", "", "postgres data-source-name")
	must(cmd.MarkFlagRequired("dsn"))
	cmd.Flags().StringVarP(&file, "file", "f", "", "file to read the tokens from. will read from stdin if omitted")
	cmd.Flags().IntVarP(&bs, "bs", "b", 10000, "batch size used for bulk import")
	return cmd
}

var tokenRe = regexp.MustCompile(`^[a-z]{7}$`)

func importTokens(r io.Reader, dsn string, bs int, outErr io.Writer) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS tokens (token CHAR(7) PRIMARY KEY)")
	if err != nil {
		return fmt.Errorf("could not create/ensure table: %w", err)
	}

	q := newImportTokensQuery(bs)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		token := scanner.Text()
		if !tokenRe.MatchString(token) {
			fmt.Fprintf(outErr, "token %s is invalid, skipping\n", token)
			continue
		}
		err = q.AddToken(token)
		if err != nil {
			return fmt.Errorf("could not add token to query: %w", err)
		}
		if q.Full() {
			err = q.Exec(db)
			if err != nil {
				return fmt.Errorf("could not execute query: %w", err)
			}
		}
	}
	err = q.Exec(db)
	if err != nil {
		return fmt.Errorf("could not execute query: %w", err)
	}
	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("could not scan tokens: %w", err)
	}

	return nil
}

type importTokensQuery struct {
	sb   strings.Builder
	args []interface{}
	bs   int
}

func newImportTokensQuery(bs int) *importTokensQuery {
	var q importTokensQuery
	q.sb.WriteString("INSERT INTO tokens (token) VALUES ")
	q.args = make([]interface{}, 0, q.bs)
	q.bs = bs
	return &q
}

func (q *importTokensQuery) AddToken(token string) error {
	l := len(q.args)
	if l == q.bs {
		return fmt.Errorf("query is full")
	}
	if l == 0 {
		q.sb.WriteString(fmt.Sprintf("($%d)", l+1))
	} else {
		q.sb.WriteString(fmt.Sprintf(",($%d)", l+1))
	}
	q.args = append(q.args, token)
	return nil
}

func (q *importTokensQuery) Full() bool {
	return len(q.args) == q.bs
}

func (q *importTokensQuery) Exec(db *sql.DB) error {
	if len(q.args) == 0 {
		return nil
	}
	q.sb.WriteString(" ON CONFLICT DO NOTHING")
	_, err := db.Exec(q.sb.String(), q.args...)
	if err != nil {
		return err
	}
	q.sb.Reset()
	q.args = make([]interface{}, 0, q.bs)
	q.sb.WriteString("INSERT INTO tokens (token) VALUES ")
	return nil
}
