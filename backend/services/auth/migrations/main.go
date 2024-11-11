package migrations

import (
	"database/sql"
	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/logger"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func StartMigration() {
	dbstring := config.MyConfig.Database.ConnectionString
	db, err := sql.Open("pgx", dbstring)
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
