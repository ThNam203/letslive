package main

import (
	"context"
	"net"
	"os"

	// TODO: add swagger
	//_ "sen1or/lets-live/auth/docs"

	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/logger"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	dbConn, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
	if err != nil {
		logger.Panicf("unable to connect to database: %v\n", "err", err)
	}
	defer dbConn.Close(context.Background())

	serverURL := net.JoinHostPort(config.AUTH_SERVER_HOST, config.AUTH_SERVER_PORT)
	server := NewAPIServer(dbConn, serverURL)
	server.ListenAndServe(false)

	select {}
}
