package main

import (
	"database/sql"
	"io/ioutil"
	"os"
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

	f, err := ioutil.TempFile("", "tokens.txt")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(`jriwhbo
xwqpvnz
apvvirw
abc12
gjzdxeg
phnflau
apvvirw
gjzdxeg
`)
	require.Nil(t, err)

	cmd := importCmd()
	err = cmd.Flags().Set("dsn", dsn)
	require.Nil(t, err)
	err = cmd.Flags().Set("file", f.Name())
	require.Nil(t, err)

	err = cmd.Execute()
	require.Nil(t, err)

	db, err := sql.Open("postgres", dsn)
	require.Nil(t, err)
	defer db.Close()

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
		"gjzdxeg",
		"phnflau",
	}, tokens)
}
