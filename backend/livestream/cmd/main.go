package main

import (
	"context"
	"fmt"

	cfg "sen1or/lets-live/livestream/config"
	"sen1or/lets-live/livestream/handlers"
	"sen1or/lets-live/livestream/repositories"
	"sen1or/lets-live/livestream/services"
	"sen1or/lets-live/livestream/utils"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	godotenv.Load("livestream/.env")
	logger.Init(logger.LogLevel(logger.Debug))
	config := cfg.RetrieveConfig()

	logger.Debugf("connection string: %s", config.Database.ConnectionString)
	utils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

	// for consul service discovery
	registry, err := discovery.NewConsulRegistry(config.Registry.RegistryService.Address)
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
		panic(1)
	}
	go RegisterToDiscoveryService(ctx, registry, config)

	dbConn := ConnectDB(ctx, config)
	defer dbConn.Close()

	server := SetupServer(dbConn, registry, *config)
	go server.ListenAndServe(false)
	select {}
}

func ConnectDB(ctx context.Context, config *cfg.Config) *pgxpool.Pool {
	dbConn, err := pgxpool.New(ctx, config.Database.ConnectionString)
	if err != nil {
		logger.Panicf("unable to connect to database: %v\n", "err", err)
	}

	return dbConn
}

func RegisterToDiscoveryService(ctx context.Context, registry discovery.Registry, config *cfg.Config) {
	serviceName := config.Service.Name
	serviceHostPort := fmt.Sprintf("%s:%d", config.Service.Hostname, config.Service.APIPort)
	serviceHealthCheckURL := fmt.Sprintf("http://%s/v1/health", serviceHostPort)
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, serviceName, instanceID, config.Registry.RegistryService.Tags); err != nil {
		logger.Panicf("failed to register server: %s", err)
	}

	ctx, _ = context.WithCancel(ctx)

	<-ctx.Done()

	if err := registry.Deregister(ctx, serviceName, instanceID); err != nil {
		logger.Errorf("failed to deregister service: %s", err)
	}
}

func SetupServer(dbConn *pgxpool.Pool, registry discovery.Registry, cfg cfg.Config) *APIServer {
	var livestreamRepo = repositories.NewLivestreamRepository(dbConn)
	var livestreamService = services.NewLivestreamService(livestreamRepo)
	var livestreamHandler = handlers.NewLivestreamHandler(*livestreamService)
	return NewAPIServer(livestreamHandler, cfg)
}
