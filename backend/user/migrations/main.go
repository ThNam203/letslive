package migrations

import (
	"database/sql"
	"sen1or/lets-live/pkg/logger"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func StartMigration(connectionString string) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("goose: failed to close db %v\n", err)
		}
	}()

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
