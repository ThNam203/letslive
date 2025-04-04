package utils

import (
	"database/sql"
	"sen1or/letslive/livestream/pkg/logger"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func StartMigration(connectionString string, migrationPath string) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Panicf("failed to open connection to db (%s): %s", connectionString, err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("goose: failed to close db %v\n", err)
		}
	}()

	if err := goose.Up(db, migrationPath); err != nil {
		panic(err)
	}
}
