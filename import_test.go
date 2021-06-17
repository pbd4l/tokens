package main

import (
	"bytes"
	"database/sql"
	"os"
	"regexp"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pbd4l/tokens/testutils"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	pg, err := testutils.NewPostgresContainer("tokens")
	require.NoError(t, err)
	t.Cleanup(func() {
		err = pg.Terminate()
		require.NoError(t, err)
	})

	dsn, err := pg.Dsn()
	require.NoError(t, err)

	cmd := importCmd()
	err = os.Setenv("TOKENS_DSN", dsn)
	require.NoError(t, err)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	t.Run("simple import", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.NoError(t, err)

		var stdin bytes.Buffer
		stdin.WriteString(`jriwhbo
xwqpvnz
apvvirw
`)
		cmd.SetIn(&stdin)

		err = cmd.Execute()
		require.NoError(t, err)

		rows, err := db.Query("SELECT token FROM tokens")
		require.NoError(t, err)
		count := 0
		tokens := make(map[string]bool)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.NoError(t, err)
			count++
			tokens[token] = true
		}
		require.Equal(t, 3, count)
		require.Equal(t, map[string]bool{
			"jriwhbo": true,
			"xwqpvnz": true,
			"apvvirw": true,
		}, tokens)
	})

	t.Run("import with duplicates", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.NoError(t, err)

		var stdin bytes.Buffer
		stdin.WriteString(`abiwhbo
pzqpvnz
dhvvirw
pzqpvnz
abiwhbo
abiwhbo
`)
		cmd.SetIn(&stdin)

		var stderr bytes.Buffer
		cmd.SetErr(&stderr)

		err = cmd.Execute()
		require.NoError(t, err)

		rows, err := db.Query("SELECT token FROM tokens")
		require.NoError(t, err)
		count := 0
		tokens := make(map[string]bool)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.NoError(t, err)
			count++
			tokens[token] = true
		}
		require.Equal(t, 3, count)
		require.Equal(t, map[string]bool{
			"abiwhbo": true,
			"pzqpvnz": true,
			"dhvvirw": true,
		}, tokens)

		logs, err := testutils.NewLogs(&stderr)
		require.NoError(t, err)
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "abiwhbo" appears 3 times, only importing once`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "pzqpvnz" appears 2 times, only importing once`)))
	})

	t.Run("import with invalid tokens", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.NoError(t, err)

		var stdin bytes.Buffer
		stdin.WriteString(`1234567
ABCDEFG
hvvi

dfiwhsz
`)
		cmd.SetIn(&stdin)

		var stderr bytes.Buffer
		cmd.SetErr(&stderr)

		err = cmd.Execute()
		require.NoError(t, err)

		rows, err := db.Query("SELECT token FROM tokens")
		require.NoError(t, err)
		count := 0
		tokens := make(map[string]bool)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.NoError(t, err)
			count++
			tokens[token] = true
		}
		require.Equal(t, 1, count)
		require.Equal(t, map[string]bool{
			"dfiwhsz": true,
		}, tokens)

		logs, err := testutils.NewLogs(&stderr)
		require.NoError(t, err)
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "1234567" is invalid, skipping`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "ABCDEFG" is invalid, skipping`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "hvvi" is invalid, skipping`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "" is invalid, skipping`)))
	})
}
