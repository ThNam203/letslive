package utils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestStartMigration(t *testing.T) {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		Env:          map[string]string{"POSTGRES_PASSWORD": "password", "POSTGRES_DB": "testdb"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatalf("failed to start PostgreSQL container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")
	connectionString := fmt.Sprintf("postgres://postgres:password@%s:%s/testdb?sslmode=disable", host, port.Port())

	migrationPath := "../migrations"
	if _, err := os.Stat(migrationPath); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("migration path does not exist")
		} else {
			t.Fatalf("failed to get migration path detail: %s", err)
		}
	}

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	t.Run("up_migration", func(t *testing.T) {
		if err := goose.Up(db, migrationPath); err != nil {
			t.Fatalf("up migration failed: %v", err)
		}
	})

	// rollback
	t.Run("down_migration", func(t *testing.T) {
		if err := goose.Down(db, migrationPath); err != nil {
			t.Fatalf("down migration failed: %v", err)
		}
	})
}
