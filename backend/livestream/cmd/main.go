package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sen1or/letslive/livestream/api"
	cfg "sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/handlers"
	"sen1or/letslive/livestream/pkg/discovery"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/repositories"
	"sen1or/letslive/livestream/services"
	"sen1or/letslive/livestream/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName    = "livestream_service"
	configProfile        = os.Getenv("CONFIG_SERVER_PROFILE")
	configReloadInterval = 30 * time.Second

	discoveryBaseDelay = 1 * time.Second
	discoveryMaxDelay  = 1 * time.Minute

	shutdownTimeout = 15 * time.Second
)

func main() {
	logger.Init(logger.LogLevel(logger.Debug))
	// for consul service discovery
	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
		panic(1)
	}

	cfgManager, err := cfg.NewConfigManager(registry, configServiceName, configProfile, configReloadInterval)
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config := cfgManager.GetConfig()
	utils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)
	go RegisterToDiscoveryService(ctx, registry, config)

	dbConn := ConnectDB(ctx, config)
	defer dbConn.Close()

	server := SetupServer(dbConn, registry, *config)
	go func() {
		logger.Infof("starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		// ListenAndServe should ideally block until an error occurs (e.g., server stopped)
		server.ListenAndServe(false)
		stop() // trigger shutdown if server fails unexpectedly
	}()

	logger.Infof("server started.")
	<-ctx.Done() // block here until SIGINT/SIGTERM is received (ctx from signal.NotifyContext)

	// initiate graceful shutdown
	logger.Infof("shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout) // Adjust timeout as needed
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		if err == context.DeadlineExceeded {
			logger.Errorf("server shutdown timed out after 15 seconds.")
		}
	} else {
		logger.Infof("server shutdown initiated successfully via main.")
	}

	logger.Infof("service shut down complete.")
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

	currentDelay := discoveryBaseDelay

	logger.Infof("attempting to register service '%s' instance '%s' [%s] with discovery service...", serviceName, instanceID, serviceHostPort)

	for {
		err := registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, serviceName, instanceID, nil) // Pass metadata if needed
		if err == nil {
			logger.Infof("successfully registered service '%s' instance '%s'", serviceName, instanceID)
			break
		}

		logger.Errorf("failed to register service '%s' instance '%s': %v - retrying in %v...", serviceName, instanceID, err, currentDelay)

		// Wait for the current delay duration, but also listen for context cancellation
		timer := time.NewTimer(currentDelay)
		select {
		case <-ctx.Done():
			// context was cancelled during the wait
			logger.Warnf("registration attempt cancelled for service '%s' instance '%s' due to context cancellation: %v", serviceName, instanceID, ctx.Err())
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

	// wait for the application context to be cancelled (e.g., shutdown signal)
	// before attempting to deregister.
	logger.Infof("service '%s' instance '%s' registered. Waiting for context cancellation signal to deregister...", serviceName, instanceID)
	<-ctx.Done() // Wait for shutdown signal

	logger.Infof("context cancelled - attempting to deregister service '%s' instance '%s'...", serviceName, instanceID)

	// Create a new short-lived context for the deregistration call.
	// The original context `ctx` is already cancelled, so using it might cause immediate failure.
	// Use context.Background() as the parent to ensure it's not tied to the cancelled context.
	deregisterCtx, cancelDeregister := context.WithTimeout(context.Background(), 5*time.Second) // 5-second timeout for deregistration
	defer cancelDeregister()

	if err := registry.Deregister(deregisterCtx, serviceName, instanceID); err != nil {
		logger.Errorf("failed to deregister service '%s' instance '%s': %v", serviceName, instanceID, err)
	} else {
		logger.Infof("successfully deregistered service '%s' instance '%s'", serviceName, instanceID)
	}
}

func SetupServer(dbConn *pgxpool.Pool, registry discovery.Registry, cfg cfg.Config) *api.APIServer {
	var livestreamRepo = repositories.NewLivestreamRepository(dbConn)
	var livestreamService = services.NewLivestreamService(livestreamRepo)
	var livestreamHandler = handlers.NewLivestreamHandler(*livestreamService)
	return api.NewAPIServer(livestreamHandler, cfg)
}
