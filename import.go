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
	"github.com/spf13/viper"
)

func importCmd() *cobra.Command {
	var file string
	var bs int

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

			dsn := viper.GetString("dsn")
			if dsn == "" {
				return fmt.Errorf("missing postgres data source name")
			}

			return importTokens(ctx, r, dsn, bs)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "file to read the tokens from. will read from stdin if omitted")
	cmd.Flags().IntVarP(&bs, "bs", "b", 1000, "batch size used for bulk import")

	viper.SetEnvPrefix("tokens")

	cmd.Flags().String("dsn", "", "postgres data source name (env TOKENS_DSN)")
	must(viper.BindPFlag("dsn", cmd.Flags().Lookup("dsn")))
	must(viper.BindEnv("dsn"))

	return cmd
}

var tokenRe = regexp.MustCompile(`^[a-z]{7}$`)

func importTokens(ctx context.Context, r io.Reader, dsn string, bs int) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS tokens (token CHAR(7) PRIMARY KEY)")
	if err != nil {
		return fmt.Errorf("could not create/ensure table: %w", err)
	}

	q := newInsertTokensQuery(bs)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		token := scanner.Text()
		if !tokenRe.MatchString(token) {
			log.Printf("token \"%s\" is invalid, skipping\n", token)
			continue
		}
		err = q.AddToken(token)
		if err != nil {
			return fmt.Errorf("could not add token to query: %w", err)
		}
		if q.Full() {
			err = q.Exec(tx)
			if err != nil {
				return fmt.Errorf("could not execute query: %w", err)
			}
		}
	}
	err = q.Exec(tx)
	if err != nil {
		return fmt.Errorf("could not execute query: %w", err)
	}
	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("could not scan tokens: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

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
	q.sb.WriteString(" ON CONFLICT DO NOTHING")
	_, err := tx.Exec(q.sb.String(), q.args...)
	if err != nil {
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
