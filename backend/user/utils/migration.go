package utils

import (
	"context"
	"database/sql"
	"sen1or/letslive/user/pkg/logger"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func StartMigration(connectionString string, migrationPath string) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Panicf(context.TODO(), "failed to start migration: %s", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf(context.TODO(), "goose: failed to close db %v\n", err)
		}
	}()

	if err := goose.Up(db, migrationPath); err != nil {
		logger.Panicf(context.TODO(), "failed to migrate on path %s: %s", migrationPath, err)
	}
}
