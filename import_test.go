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
	require.Nil(t, err)
	t.Cleanup(func() {
		err = pg.Terminate()
		require.Nil(t, err)
	})

	dsn, err := pg.Dsn()
	require.Nil(t, err)

	cmd := importCmd()
	err = os.Setenv("TOKENS_DSN", dsn)
	require.Nil(t, err)

	db, err := sql.Open("postgres", dsn)
	require.Nil(t, err)
	defer db.Close()

	t.Run("simple import", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.Nil(t, err)

		var stdin bytes.Buffer
		stdin.WriteString(`jriwhbo
xwqpvnz
apvvirw
`)
		cmd.SetIn(&stdin)

		err = cmd.Execute()
		require.Nil(t, err)

		tokens := make([]string, 0)
		rows, err := db.Query("SELECT token FROM tokens")
		require.Nil(t, err)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.Nil(t, err)
			tokens = append(tokens, token)
		}
		require.Equal(t, []string{
			"jriwhbo",
			"xwqpvnz",
			"apvvirw",
		}, tokens)
	})

	t.Run("import with duplicates", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.Nil(t, err)

		var stdin bytes.Buffer
		stdin.WriteString(`abiwhbo
pzqpvnz
dhvvirw
pzqpvnz
abiwhbo
`)
		cmd.SetIn(&stdin)

		err = cmd.Execute()
		require.Nil(t, err)

		tokens := make([]string, 0)
		rows, err := db.Query("SELECT token FROM tokens")
		require.Nil(t, err)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.Nil(t, err)
			tokens = append(tokens, token)
		}
		require.Equal(t, []string{
			"abiwhbo",
			"pzqpvnz",
			"dhvvirw",
		}, tokens)
	})

	t.Run("import twice", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.Nil(t, err)

		var stdin bytes.Buffer
		stdin.WriteString(`bviwhbo
qxqpvnz
kkvvirw
`)
		cmd.SetIn(&stdin)

		err = cmd.Execute()
		require.Nil(t, err)

		stdin = bytes.Buffer{}
		stdin.WriteString(`bviwhbo
qxqpvnz
agvvirw
kkvvirw
`)
		cmd.SetIn(&stdin)

		err = cmd.Execute()
		require.Nil(t, err)

		tokens := make([]string, 0)
		rows, err := db.Query("SELECT token FROM tokens")
		require.Nil(t, err)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.Nil(t, err)
			tokens = append(tokens, token)
		}
		require.Equal(t, []string{
			"bviwhbo",
			"qxqpvnz",
			"kkvvirw",
			"agvvirw",
		}, tokens)
	})

	t.Run("import with invalid tokens", func(t *testing.T) {
		_, err = db.Exec("DROP TABLE IF EXISTS tokens")
		require.Nil(t, err)

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
		require.Nil(t, err)

		tokens := make([]string, 0)
		rows, err := db.Query("SELECT token FROM tokens")
		require.Nil(t, err)
		for rows.Next() {
			var token string
			err = rows.Scan(&token)
			require.Nil(t, err)
			tokens = append(tokens, token)
		}
		require.Equal(t, []string{
			"dfiwhsz",
		}, tokens)

		logs, err := testutils.NewLogs(&stderr)
		require.Nil(t, err)
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "1234567" is invalid, skipping`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "ABCDEFG" is invalid, skipping`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "hvvi" is invalid, skipping`)))
		require.True(t, logs.ContainsMatch(regexp.MustCompile(`token "" is invalid, skipping`)))
	})
}
