package utils

import (
	"context"
	"database/sql"
	"sen1or/letslive/livestream/pkg/logger"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func StartMigration(connectionString string, migrationPath string) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Panicf(context.TODO(), "failed to open connection to db (%s): %s", connectionString, err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf(context.TODO(), "goose: failed to close db %v\n", err)
		}
	}()

	if err := goose.Up(db, migrationPath); err != nil {
		panic(err)
	}
}
