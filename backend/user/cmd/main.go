package main

import (
	"context"
	"fmt"
	"os"

	cfg "sen1or/letslive/user/config"
	"sen1or/letslive/user/handlers"
	"sen1or/letslive/user/pkg/discovery"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/repositories"
	"sen1or/letslive/user/services"
	"sen1or/letslive/user/utils"

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
	var userRepo = repositories.NewUserRepository(dbConn)
	var livestreamInfoRepo = repositories.NewLivestreamInformationRepository(dbConn)
	var followRepo = repositories.NewFollowRepository(dbConn)

	minioService := services.NewMinIOService(context.Background(), cfg.MinIO)
	var userService = services.NewUserService(userRepo, livestreamInfoRepo, *minioService)
	var livestreamInfoService = services.NewLivestreamInformationService(livestreamInfoRepo)
	var followService = services.NewFollowService(followRepo)

	var userHandler = handlers.NewUserHandler(*userService)
	var livestreamInfoHandler = handlers.NewLivestreamInformationHandler(*livestreamInfoService, *minioService)
	var followHandler = handlers.NewFollowHandler(*followService)
	return NewAPIServer(userHandler, livestreamInfoHandler, followHandler, cfg)
}
