package utils

import (
	"database/sql"
	"os"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func StartMigration(connectionString string, migrationPath string) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("goose: failed to close db %v\n", err)
		}
	}()

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	logger.Infof("currently executing path: %s", exPath)

	if err := goose.Up(db, migrationPath); err != nil {
		panic(err)
	}
}
