package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose"
)

func init() {
	dbstring := os.Getenv("POSTGRES_URL")
	fmt.Println(dbstring)
	db, err := sql.Open("pgx", dbstring)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close db %v\n", err)
		}
	}()

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
