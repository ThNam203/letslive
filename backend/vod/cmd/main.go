package main

import (
	"context"
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
	"sen1or/letslive/vod/repositories"
	vodService "sen1or/letslive/vod/services/vod"
	vodCommentService "sen1or/letslive/vod/services/vod_comment"
	miniostorage "sen1or/letslive/vod/storage/minio"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "vod_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

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

	cfgManager, err := sharedconfig.NewConfigManager[cfg.Config](ctx, registry, configServiceName, configProfile, cfg.PostProcess)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}
	defer cfgManager.Stop()

	config := cfgManager.GetConfig()
	sharedutils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

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
	return api.NewAPIServer(vodHandler, vodCommentHandler, cfg, dbConn)
}
