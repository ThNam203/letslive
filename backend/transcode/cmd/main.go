package main

import (
	"context"
	"os"
	"os/signal"
	cfg "sen1or/letslive/transcode/config"
	livestreamgateway "sen1or/letslive/transcode/gateway/livestream/http"
	usergateway "sen1or/letslive/transcode/gateway/user/http"
	"sen1or/letslive/transcode/rtmp"
	miniostorage "sen1or/letslive/transcode/storage/minio"
	"sen1or/letslive/transcode/watcher"
	miniowatcher "sen1or/letslive/transcode/watcher/minio"
	"sen1or/letslive/transcode/webserver"
	"sen1or/letslive/transcode/worker"
	"sync"
	"syscall"
	"time"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	sharedutils "sen1or/letslive/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "transcode_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	gracefulShutdownTimeout = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Init(logger.LogLevel(logger.Debug))
	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf(ctx, "failed to get a new registry")
	}

	cfgManager, err := sharedconfig.NewConfigManager[cfg.Config](ctx, registry, configServiceName, configProfile, cfg.PostProcess)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}

	config := cfgManager.GetConfig()

	// service discovery
	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go sharedutils.RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config.Service.Hostname, config.Webserver.Port)

	setupHLSFolders(config.Transcode)

	// TODO: fix this, we need a webserver built into the transcode server (not the pkg/webserver, use nginx instead)
	allowedSuffixes := [2]string{".ts", ".m3u8"}
	MyWebServer := webserver.NewWebServer(config.Webserver.Port, allowedSuffixes[:], config.Transcode.PublicHLSPath)
	go MyWebServer.ListenAndServe()

	var vodHandler watcher.VODHandler

	if !config.MinIO.Enabled {
		logger.Warnf(ctx, "minio is forced to be enable, we are ignoring minio.enabled")
	}

	vodHandler = miniowatcher.GetMinIOVODStrategy()
	minioStorage := miniostorage.NewMinIOStorage(ctx, config.MinIO)
	minioWatcherStrategy := miniowatcher.NewMinIOFileWatcherStrategy(vodHandler, minioStorage, *config)
	watcher := watcher.NewFFMpegFileWatcher(config.Transcode.PrivateHLSPath, minioWatcherStrategy)
	go watcher.Watch(ctx)

	userGateway := usergateway.NewUserGateway(registry)
	livestreamGateway := livestreamgateway.NewLivestreamGateway(registry)

	// Initialize transcode worker for uploaded video processing
	var transcodeWorker *worker.TranscodeWorker
	if config.Database.ConnectionString != "" {
		poolConfig, parseErr := pgxpool.ParseConfig(config.Database.ConnectionString)
		if parseErr != nil {
			logger.Errorf(ctx, "failed to parse database connection string: %v", parseErr)
		}

		if poolConfig != nil {
			poolConfig.MaxConns = 10
			poolConfig.MinConns = 1
			poolConfig.MaxConnLifetime = 30 * time.Minute
			poolConfig.MaxConnIdleTime = 5 * time.Minute
			poolConfig.HealthCheckPeriod = 30 * time.Second

			dbConn, dbErr := pgxpool.NewWithConfig(ctx, poolConfig)
			if dbErr != nil {
				logger.Errorf(ctx, "failed to connect to database for transcode worker: %v", dbErr)
			} else {
				rawMinioClient := worker.NewRawMinIOClient(config.MinIO)
				transcodeWorker = worker.NewTranscodeWorker(
					dbConn,
					minioStorage,
					rawMinioClient,
					config.MinIO.BucketName,
					config,
					livestreamGateway,
				)
				go transcodeWorker.Start(ctx)
				logger.Infof(ctx, "transcode worker started")
			}
		}
	}

	// TODO: find a way to remove the vodHandler from the rtmp, or change the design or config
	//Use kafka
	rtmpServer := rtmp.NewRTMPServer(
		rtmp.RTMPServerConfig{Context: ctx, Port: config.RTMP.Port, Registry: &registry, Config: *config, VODHandler: vodHandler},
		userGateway,
		livestreamGateway,
	)
	go rtmpServer.Start()
	<-ctx.Done()

	logger.Infof(ctx, "starting coordinated shutdown...")

	// Create a shutdown context with timeout
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancelShutdown()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		MyWebServer.Shutdown(shutdownCtx)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		rtmpServer.Shutdown(shutdownCtx)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		watcher.Shutdown()
		wg.Done()
	}()

	if transcodeWorker != nil {
		wg.Add(1)
		go func() {
			transcodeWorker.Shutdown()
			wg.Done()
		}()
	}

	wg.Add(1)
	go (func() {
		sharedutils.DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
		wg.Done()
	})()

	wg.Wait()
	logger.Infof(ctx, "service shutdown complete.")
}

func setupHLSFolders(cfg cfg.Transcode) {
	if err := os.MkdirAll(cfg.PublicHLSPath, 0777); err != nil {
		logger.Panicf(context.TODO(), "failed to create public hls folder: %s", err)
	}

	if err := os.MkdirAll(cfg.PrivateHLSPath, 0777); err != nil {
		logger.Panicf(context.TODO(), "failed to create private hls folder: %s", err)
	}
}
