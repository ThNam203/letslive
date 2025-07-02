package utils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.NoError(t, err, "failed to start PostgreSQL container")
	defer pgContainer.Terminate(ctx)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)
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

	t.Run("up_migration_and_verify", func(t *testing.T) {
		var panicErr any
		func() {
			defer func() {
				panicErr = recover()
			}()
			// Call the function under test
			StartMigration(connectionString, migrationPath)
		}()

		require.Nil(t, panicErr, "StartMigration panicked unexpectedly: %v", panicErr)
	})

	// rollback
	t.Run("down_migration", func(t *testing.T) {
		if err := goose.Down(db, migrationPath); err != nil {
			t.Fatalf("down migration failed: %v", err)
		}
	})

	t.Run("down_migration_and_verify", func(t *testing.T) {
		err := goose.Down(db, migrationPath)
		require.NoError(t, err, "goose.Down migration failed")

		// verification after DOWN
		expectedTable := "your_table_name_from_migration"
		exists, err := tableExists(db, expectedTable)
		require.NoError(t, err, "failed to check if table '%s' exists after down migration", expectedTable)
		assert.False(t, exists, "table '%s' should NOT exist after running migration down, but it still does", expectedTable)
	})
}

// countUserTables counts tables in the 'public' schema, excluding the goose version table.
func countUserTables(db *sql.DB, gooseTableName string) (int, error) {
	// Query to count tables in 'public' schema, excluding the specified goose table
	// Adjust 'public' if your migrations target a different schema.
	query := `
        SELECT COUNT(*)
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name != $1;
    `
	var count int
	err := db.QueryRow(query, gooseTableName).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user tables: %w", err)
	}
	return count, nil
}
