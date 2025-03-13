package main

import (
	"context"
	"fmt"
	"os"

	cfg "sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/handlers"
	"sen1or/letslive/livestream/pkg/discovery"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/repositories"
	"sen1or/letslive/livestream/services"
	"sen1or/letslive/livestream/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	logger.Init(logger.LogLevel(logger.Debug))
	// for consul service discovery
	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
		panic(1)
	}

	config := cfg.RetrieveConfig(registry)

	utils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)
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
	if err := registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, serviceName, instanceID, nil); err != nil {
		logger.Panicf("failed to register server: %s", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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
