package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {
	var file string
	var bs int
	var dsn string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import tokens to a postgres database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetOutput(cmd.ErrOrStderr()) // for testing

			ctx, _ := signal.NotifyContext(cmd.Context(), syscall.SIGINT)

			var r io.Reader = cmd.InOrStdin()
			if file != "" {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("could not open tokens file: %w", err)
				}
				defer f.Close()
				r = f
			}

			if dsn == "" {
				dsn = os.Getenv("TOKENS_DSN")
			}
			if dsn == "" {
				return fmt.Errorf("missing postgres data source name")
			}

			return importTokens(ctx, r, dsn, bs)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "file to read the tokens from. will read from stdin if omitted")
	cmd.Flags().IntVarP(&bs, "bs", "b", 1000, "batch size used for bulk import")
	cmd.Flags().StringVar(&dsn, "dsn", "", "postgres data source name (env TOKENS_DSN)")
	return cmd
}

var tokenRegex = regexp.MustCompile(`^[a-z]{7}$`)

func importTokens(ctx context.Context, r io.Reader, dsn string, bs int) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}
	defer db.Close()

	// read tokens into a map first, so we can determine non-unique tokens.
	// note: non-unique here does not account for tokens already in the database, only tokens to be imported.
	// note: a map is not ordered, so currently we are not certain which order the tokens will be inserted.
	tokens := make(map[string]uint8)
	s := bufio.NewScanner(r)
	for s.Scan() {
		token := s.Text()
		if !tokenRegex.MatchString(token) {
			log.Printf("token \"%s\" is invalid, skipping\n", token)
			continue
		}
		tokens[token]++
	}
	if err = s.Err(); err != nil {
		return fmt.Errorf("could not scan tokens: %w", err)
	}

	// use a transaction to ensure the import either succeeds or fails entirely.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	// ensure the tokens table exists.
	// tokens are 7 characters and must be unique, so we use CHAR(7) PRIMARY KEY.
	// (consider moving to separate migration command?)
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS tokens (token CHAR(7) PRIMARY KEY)")
	if err != nil {
		return fmt.Errorf("could not create/ensure table: %w", err)
	}

	// insert into the database
	q := newInsertTokensQuery(bs)
	for token, count := range tokens {
		if count > 1 {
			log.Printf("token \"%s\" appears %d times, only importing once", token, count)
		}
		if err = q.AddToken(token); err != nil {
			return fmt.Errorf("could not add token to query: %w", err)
		}
		if q.Full() {
			if err = q.Exec(tx); err != nil {
				return fmt.Errorf("could not execute query: %w", err)
			}
		}
	}
	if err = q.Exec(tx); err != nil {
		return fmt.Errorf("could not execute query: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// insertTokensQuery builds a bulk insert query for tokens
type insertTokensQuery struct {
	sb   strings.Builder
	args []interface{}
	ms   int
}

func newInsertTokensQuery(maxSize int) *insertTokensQuery {
	var q insertTokensQuery
	q.sb.WriteString("INSERT INTO tokens (token) VALUES ")
	q.args = make([]interface{}, 0, maxSize)
	q.ms = maxSize
	return &q
}

func (q *insertTokensQuery) Exec(tx *sql.Tx) error {
	if len(q.args) == 0 {
		return nil
	}
	// ON CONFLICT DO NOTHING to skip when inserting an already inserted token
	q.sb.WriteString(" ON CONFLICT DO NOTHING")
	if _, err := tx.Exec(q.sb.String(), q.args...); err != nil {
		return err
	}
	q.sb.Reset()
	q.sb.WriteString("INSERT INTO tokens (token) VALUES ")
	q.args = make([]interface{}, 0, q.ms)
	return nil
}

func (q *insertTokensQuery) AddToken(token string) error {
	l := len(q.args)
	if l == q.ms {
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

func (q *insertTokensQuery) Full() bool {
	return len(q.args) == q.ms
}
