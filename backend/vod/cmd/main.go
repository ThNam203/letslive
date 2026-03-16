package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/vod/api"
	cfg "sen1or/letslive/vod/config"
	usergatewayhttp "sen1or/letslive/vod/gateway/user/http"
	vodHandler "sen1or/letslive/vod/handlers/vod"
	vodCommentHandler "sen1or/letslive/vod/handlers/vod_comment"
	"sen1or/letslive/vod/pkg/discovery"
	"sen1or/letslive/vod/pkg/logger"
	"sen1or/letslive/vod/pkg/tracer"
	"sen1or/letslive/vod/repositories"
	miniostorage "sen1or/letslive/vod/storage/minio"
	vodService "sen1or/letslive/vod/services/vod"
	vodCommentService "sen1or/letslive/vod/services/vod_comment"
	"sen1or/letslive/vod/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "vod_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	discoveryBaseDelay = 1 * time.Second
	discoveryMaxDelay  = 1 * time.Minute

	shutdownTimeout = 15 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Init(logger.LogLevel(logger.Debug))

	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf(ctx, "failed to start discovery mechanism: %s", err)
	}

	cfgManager, err := cfg.NewConfigManager(ctx, registry, configServiceName, configProfile)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}
	defer cfgManager.Stop()

	config := cfgManager.GetConfig()
	utils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

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
		server.ListenAndServe(ctx, false)
		stop()
	}()

	logger.Infof(ctx, "server started.")
	<-ctx.Done()

	logger.Infof(ctx, "shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, shutdownTimeout)
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
		err := registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, serviceName, instanceId, nil)
		if err == nil {
			logger.Infof(ctx, "successfully registered service '%s' instance '%s'", serviceName, instanceId)
			break
		}

		logger.Errorf(ctx, "failed to register service '%s' instance '%s': %v - retrying in %v...", serviceName, instanceId, err, currentDelay)

		timer := time.NewTimer(currentDelay)
		select {
		case <-ctx.Done():
			logger.Warnf(ctx, "registration attempt cancelled for service '%s' instance '%s' due to context cancellation: %v", serviceName, instanceId, ctx.Err())
			timer.Stop()
			return
		case <-timer.C:
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
	var vodRepo = repositories.NewVODRepository(dbConn)
	var vodCommentRepo = repositories.NewVODCommentRepository(dbConn)
	var vodCommentLikeRepo = repositories.NewVODCommentLikeRepository(dbConn)
	var transcodeJobRepo = repositories.NewTranscodeJobRepository(dbConn)

	var userGateway = usergatewayhttp.NewUserGateway(registry)

	var minio = miniostorage.NewMinIOStorage(ctx, cfg.MinIO)

	var vodService = vodService.NewVODService(vodRepo, transcodeJobRepo, minio)
	var vodCommentService = vodCommentService.NewVODCommentService(vodCommentRepo, vodCommentLikeRepo, vodRepo, userGateway, dbConn)

	var vodHandler = vodHandler.NewVODHandler(vodService)
	var vodCommentHandler = vodCommentHandler.NewVODCommentHandler(vodCommentService)
	return api.NewAPIServer(vodHandler, vodCommentHandler, cfg)
}
