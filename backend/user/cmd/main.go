package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/user/api"
	cfg "sen1or/letslive/user/config"
	"sen1or/letslive/user/handlers/follow"
	"sen1or/letslive/user/handlers/livestream_information"
	notificationhandler "sen1or/letslive/user/handlers/notification"
	"sen1or/letslive/user/handlers/user"
	"sen1or/letslive/user/repositories"
	"sen1or/letslive/user/services"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "user_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	shutdownTimeout            = 15 * time.Second
	discoveryDeregisterTimeout = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Init(logger.LogLevel(logger.Debug))

	// for consul service discovery
	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf(ctx, "failed to start discovery mechanism: %s", err)
	}

	cfgManager, err := sharedconfig.NewConfigManager[cfg.Config](ctx, registry, configServiceName, configProfile, cfg.PostProcess)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}

	config := cfgManager.GetConfig()

	sharedutils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

	// service discovery
	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go sharedutils.RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config.Service.Hostname, config.Service.APIPort)

	otelShutdownFunc, err := tracer.SetupOTelSDK(ctx, *config)
	if err != nil {
		logger.Panicf(ctx, "failed to setup otel sdk: %v", err)
	}

	dbConn := sharedutils.ConnectDB(ctx, config.Database.ConnectionString)
	defer dbConn.Close()

	server := SetupServer(ctx, dbConn, registry, config)
	go func() {
		logger.Infof(ctx, "starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		// ListenAndServe should ideally block until an error occurs (e.g., server stopped)
		server.ListenAndServe(ctx, false)
		stop() // trigger shutdown if server fails unexpectedly
	}()

	logger.Infof(ctx, "server started.")
	<-ctx.Done() // block here until SIGINT/SIGTERM is received (ctx from signal.NotifyContext)

	logger.Infof(ctx, "shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, shutdownTimeout) // Adjust timeout as needed
	defer cancelShutdown()

	var shutdownWg sync.WaitGroup

	shutdownWg.Add(1)
	go (func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			if err == context.DeadlineExceeded {
				logger.Errorf(shutdownCtx, "server shutdown timed out.")
			}
		}
		shutdownWg.Done()
	})()

	shutdownWg.Add(1)
	go (func() {
		sharedutils.DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
		shutdownWg.Done()
	})()

	shutdownWg.Add(1)
	go (func() {
		otelShutdownFunc(shutdownCtx)
		shutdownWg.Done()
	})()

	shutdownWg.Wait()
	logger.Infof(shutdownCtx, "service shut down complete.")
}

func SetupServer(ctx context.Context, dbConn *pgxpool.Pool, registry discovery.Registry, cfg *cfg.Config) *api.APIServer {
	var userRepo = repositories.NewUserRepository(dbConn)
	var livestreamInfoRepo = repositories.NewLivestreamInformationRepository(dbConn)
	var followRepo = repositories.NewFollowRepository(dbConn)
	var notificationRepo = repositories.NewNotificationRepository(dbConn)

	minioService := services.NewMinIOService(ctx, cfg.MinIO)
	var userService = services.NewUserService(dbConn, userRepo, livestreamInfoRepo, notificationRepo, followRepo, *minioService)
	var livestreamInfoService = services.NewLivestreamInformationService(livestreamInfoRepo)
	var followService = services.NewFollowService(followRepo)
	var notificationService = services.NewNotificationService(notificationRepo)

	var userHandler = user.NewUserHandler(*userService)
	var livestreamInfoHandler = livestream_information.NewLivestreamInformationHandler(*livestreamInfoService, *minioService)
	var followHandler = follow.NewFollowHandler(*followService)
	var notifHandler = notificationhandler.NewNotificationHandler(*notificationService)
	return api.NewAPIServer(userHandler, livestreamInfoHandler, followHandler, notifHandler, cfg, dbConn)
}
