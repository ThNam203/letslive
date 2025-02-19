package main

import (
	"context"
	"fmt"

	cfg "sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/controllers"
	"sen1or/lets-live/auth/handlers"
	"sen1or/lets-live/auth/repositories"
	"sen1or/lets-live/auth/types"
	"sen1or/lets-live/auth/utils"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/pkg/logger"

	usergateway "sen1or/lets-live/auth/gateway/user/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	godotenv.Load("auth/.env")
	logger.Init(logger.LogLevel(logger.Debug))
	config := cfg.RetrieveConfig()
	utils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

	// for consul service discovery
	registry, err := discovery.NewConsulRegistry(config.Registry.RegistryService.Address)
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
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

	ctx, cancel := context.WithCancel(ctx)

	<-ctx.Done()

	if err := registry.Deregister(ctx, serviceName, instanceID); err != nil {
		logger.Errorf("failed to deregister service: %s", err)
	}

	cancel()
}

func SetupServer(dbConn *pgxpool.Pool, registry discovery.Registry, cfg cfg.Config) *APIServer {
	var userRepo = repositories.NewAuthRepository(dbConn)
	var refreshTokenRepo = repositories.NewRefreshTokenRepository(dbConn)
	var verifyTokenRepo = repositories.NewVerifyTokenRepo(dbConn)

	var authCtrl = controllers.NewAuthController(userRepo)
	var tokenCtrl = controllers.NewTokenController(refreshTokenRepo, types.TokenControllerConfig(cfg.Tokens))
	var verifyTokenCtrl = controllers.NewVerifyTokenController(verifyTokenRepo)
	userGateway := usergateway.NewUserGateway(registry)
	var authHandler = handlers.NewAuthHandler(tokenCtrl, authCtrl, verifyTokenCtrl, cfg.Verification.Gateway, userGateway)
	return NewAPIServer(authHandler, registry, cfg)
}
