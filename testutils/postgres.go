package testutils

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer is a postgres container
type PostgresContainer struct {
	testcontainers.Container
	ctx    context.Context
	dbname string
}

// NewPostgresContainer creates & starts a new postgres container
func NewPostgresContainer(dbname string) (*PostgresContainer, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:latest",
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       dbname,
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
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{container, ctx, dbname}, nil
}

// Terminate terminates the postgres container
func (pg *PostgresContainer) Terminate() error {
	return pg.Container.Terminate(pg.ctx)
}

// Dsn returns the postgres data source name
func (pg *PostgresContainer) Dsn() (string, error) {
	host, err := pg.Host(pg.ctx)
	if err != nil {
		return "", err
	}
	port, err := pg.MappedPort(pg.ctx, nat.Port("5432/tcp"))
	if err != nil {
		return "", err
	}
	template := "host=%s port=%d user=user password=password dbname=%s sslmode=disable"
	return fmt.Sprintf(template, host, port.Int(), pg.dbname), nil
}
