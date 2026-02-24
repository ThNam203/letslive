package main

import (
	"context"
	"fmt"
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
	"sen1or/letslive/user/pkg/discovery"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/repositories"
	"sen1or/letslive/user/services"
	"sen1or/letslive/user/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "user_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	discoveryBaseDelay = 1 * time.Second
	discoveryMaxDelay  = 1 * time.Minute

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

	cfgManager, err := cfg.NewConfigManager(ctx, registry, configServiceName, configProfile)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}

	config := cfgManager.GetConfig()

	utils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

	// service discovery
	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config)

	otelShutdownFunc, err := tracer.SetupOTelSDK(ctx, *config)
	if err != nil {
		logger.Panicf(ctx, "failed to setup otel sdk: %v", err)
	}

	dbConn := ConnectDB(ctx, config)
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
		DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
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

func ConnectDB(ctx context.Context, config *cfg.Config) *pgxpool.Pool {
	dbConn, err := pgxpool.New(ctx, config.Database.ConnectionString)
	if err != nil {
		logger.Panicf(ctx, "unable to connect to database: %v\n", "err", err)
	}

	return dbConn
}

func RegisterToDiscoveryService(ctx context.Context, registry discovery.Registry, serviceName, instanceId string, config *cfg.Config) {
	serviceHostPort := fmt.Sprintf("%s:%d", config.Service.Hostname, config.Service.APIPort)
	serviceHealthCheckURL := fmt.Sprintf("http://%s/v1/health", serviceHostPort)

	currentDelay := discoveryBaseDelay

	logger.Infof(ctx, "attempting to register service '%s' instance '%s' [%s] with discovery service...", serviceName, instanceId, serviceHostPort)

	for {
		err := registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, serviceName, instanceId, nil) // Pass metadata if needed
		if err == nil {
			logger.Infof(ctx, "successfully registered service '%s' instance '%s'", serviceName, instanceId)
			break
		}

		logger.Errorf(ctx, "failed to register service '%s' instance '%s': %v - retrying in %v...", serviceName, instanceId, err, currentDelay)

		// Wait for the current delay duration, but also listen for context cancellation
		timer := time.NewTimer(currentDelay)
		select {
		case <-ctx.Done():
			// context was cancelled during the wait
			logger.Warnf(ctx, "registration attempt cancelled for service '%s' instance '%s' due to context cancellation: %v", serviceName, instanceId, ctx.Err())
			timer.Stop()
			return
		case <-timer.C:
			// Timer fired, continue to the next retry attempt
		}

		currentDelay *= 2
		if currentDelay > discoveryMaxDelay {
			currentDelay = discoveryMaxDelay
		}
	}
}

func DeregisterDiscoveryService(shutdownContext context.Context, registry discovery.Registry, serviceName, instanceId string) {
	logger.Infof(shutdownContext, "attempting to deregister service")

	if err := registry.Deregister(shutdownContext, serviceName, instanceId); err != nil {
		logger.Errorf(shutdownContext, "failed to deregister service '%s' instance '%s': %v", serviceName, instanceId, err)
	} else {
		logger.Infof(shutdownContext, "successfully deregistered service '%s' instance '%s'", serviceName, instanceId)
	}
}

func SetupServer(ctx context.Context, dbConn *pgxpool.Pool, registry discovery.Registry, cfg *cfg.Config) *api.APIServer {
	var userRepo = repositories.NewUserRepository(dbConn)
	var livestreamInfoRepo = repositories.NewLivestreamInformationRepository(dbConn)
	var followRepo = repositories.NewFollowRepository(dbConn)
	var notificationRepo = repositories.NewNotificationRepository(dbConn)

	minioService := services.NewMinIOService(ctx, cfg.MinIO)
	var userService = services.NewUserService(userRepo, livestreamInfoRepo, notificationRepo, followRepo, *minioService)
	var livestreamInfoService = services.NewLivestreamInformationService(livestreamInfoRepo)
	var followService = services.NewFollowService(followRepo)
	var notificationService = services.NewNotificationService(notificationRepo)

	var userHandler = user.NewUserHandler(*userService)
	var livestreamInfoHandler = livestream_information.NewLivestreamInformationHandler(*livestreamInfoService, *minioService)
	var followHandler = follow.NewFollowHandler(*followService)
	var notifHandler = notificationhandler.NewNotificationHandler(*notificationService)
	return api.NewAPIServer(userHandler, livestreamInfoHandler, followHandler, notifHandler, cfg)
}
