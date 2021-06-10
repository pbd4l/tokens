package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestImport(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:latest",
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "tokens",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("PostgreSQL init process complete"),
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.Nil(t, err)
	defer container.Terminate(ctx) //nolint:errcheck

	host, err := container.Host(ctx)
	require.Nil(t, err)

	port, err := container.MappedPort(ctx, nat.Port("5432/tcp"))
	require.Nil(t, err)

	dsn := fmt.Sprintf("host=%s port=%d user=user password=password dbname=tokens sslmode=disable", host, port.Int())

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
apvvirw
gjzdxeg
`)
	require.Nil(t, err)

	err = Import([]string{
		"-file", f.Name(),
		"-dsn", dsn,
	})
	require.Nil(t, err)

	db, err := sql.Open("postgres", dsn)
	require.Nil(t, err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tokens").Scan(&count)
	require.Nil(t, err)
	require.Equal(t, 5, count)

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
