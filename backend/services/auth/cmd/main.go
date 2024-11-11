package main

import (
	"context"
	"net"

	// TODO: add swagger
	//_ "sen1or/lets-live/auth/docs"

	config "sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/discovery"
	logger "sen1or/lets-live/auth/logger"
	"sen1or/lets-live/auth/migrations"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()

	logger.Init()
	config.RetrieveConfig()
	migrations.StartMigration()

	// for consul service discovery
	go StartDiscovery(ctx)

	dbConn := ConnectDB(ctx)
	defer dbConn.Close(ctx)

	serverAddr := net.JoinHostPort(config.MyConfig.Service.Host, string(config.MyConfig.Service.Port))

	server := NewAPIServer(dbConn, serverAddr)
	go server.ListenAndServe(false)
	select {}
}

func ConnectDB(ctx context.Context) *pgx.Conn {
	dbConn, err := pgx.Connect(ctx, config.MyConfig.Database.ConnectionString)
	if err != nil {
		logger.Panicf("unable to connect to database: %v\n", "err", err)
	}

	return dbConn
}

func StartDiscovery(ctx context.Context) {
	registry, err := discovery.NewConsulRegistry(config.MyConfig.Registry.Address)
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
	}

	serviceName := config.MyConfig.Service.Name

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, config.MyConfig.Registry.Address, serviceName, instanceID); err != nil {
		logger.Panicf("failed to register server: %s", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	<-ctx.Done()

	if err := registry.Deregister(ctx, serviceName, instanceID); err != nil {
		logger.Errorf("failed to deregister service: %s", err)
	}

	cancel()
}
