package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	cfg "sen1or/letslive/transcode/config"
	livestreamgateway "sen1or/letslive/transcode/gateway/livestream/http"
	usergateway "sen1or/letslive/transcode/gateway/user/http"
	"sen1or/letslive/transcode/pkg/discovery"
	"sen1or/letslive/transcode/pkg/logger"
	"sen1or/letslive/transcode/rtmp"
	miniostorage "sen1or/letslive/transcode/storage/minio"
	"sen1or/letslive/transcode/watcher"
	miniowatcher "sen1or/letslive/transcode/watcher/minio"
	"sen1or/letslive/transcode/webserver"
	"sync"
	"syscall"
	"time"
)

var (
	configServiceName    = "transcode_service"
	configProfile        = os.Getenv("CONFIG_SERVER_PROFILE")
	configReloadInterval = 30 * time.Second

	discoveryBaseDelay = 1 * time.Second
	discoveryMaxDelay  = 1 * time.Minute

	gracefulShutdownTimeout = 10 * time.Second
)

func main() {
	logger.Init(logger.LogLevel(logger.Debug))
	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf("failed to get a new registry")
	}

	cfgManager, err := cfg.NewConfigManager(registry, configServiceName, configProfile, configReloadInterval)
	if err != nil {
		logger.Panicf("failed to set up config manager: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config := cfgManager.GetConfig()

	// service discovery
	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config)

	setupHLSFolders(config.Transcode)

	// TODO: fix this, we need a webserver built into the transcode server (not the pkg/webserver, use nginx instead)
	allowedSuffixes := [2]string{".ts", ".m3u8"}
	MyWebServer := webserver.NewWebServer(config.Webserver.Port, allowedSuffixes[:], config.Transcode.PublicHLSPath)
	go MyWebServer.ListenAndServe()

	var vodHandler watcher.VODHandler

	if !config.MinIO.Enabled {
		logger.Warnf("minio is forced to be enable, we are ignoring minio.enabled")
	}

	vodHandler = miniowatcher.GetMinIOVODStrategy()
	minioStorage := miniostorage.NewMinIOStorage(ctx, config.MinIO)
	minioWatcherStrategy := miniowatcher.NewMinIOFileWatcherStrategy(vodHandler, minioStorage, *config)
	watcher := watcher.NewFFMpegFileWatcher(config.Transcode.PrivateHLSPath, minioWatcherStrategy)
	go watcher.Watch(ctx)

	userGateway := usergateway.NewUserGateway(registry)
	livestreamGateway := livestreamgateway.NewLivestreamGateway(registry)

	// TODO: find a way to remove the vodHandler from the rtmp, or change the design or config
	//Use kafka
	rtmpServer := rtmp.NewRTMPServer(
		rtmp.RTMPServerConfig{Context: ctx, Port: config.RTMP.Port, Registry: &registry, Config: *config, VODHandler: vodHandler},
		userGateway,
		livestreamGateway,
	)
	go rtmpServer.Start()
	<-ctx.Done()

	logger.Infof("starting coordinated shutdown...")

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

	wg.Add(1)
	go (func() {
		DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
		wg.Done()
	})()

	wg.Wait()
	logger.Infof("service shutdown complete.")
}

func RegisterToDiscoveryService(ctx context.Context, registry discovery.Registry, serviceName, instanceId string, config *cfg.Config) {
	serviceHostPort := fmt.Sprintf("%s:%d", config.Service.Hostname, config.Webserver.Port)
	serviceHealthCheckURL := fmt.Sprintf("http://%s/v1/health", serviceHostPort)

	currentDelay := discoveryBaseDelay

	logger.Infof("attempting to register service '%s' instance '%s' [%s] with discovery service...", serviceName, instanceId, serviceHostPort)

	for {
		err := registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, serviceName, instanceId, nil) // Pass metadata if needed
		if err == nil {
			logger.Infof("successfully registered service '%s' instance '%s'", serviceName, instanceId)
			break
		}

		logger.Errorf("failed to register service '%s' instance '%s': %v - retrying in %v...", serviceName, instanceId, err, currentDelay)

		// Wait for the current delay duration, but also listen for context cancellation
		timer := time.NewTimer(currentDelay)
		select {
		case <-ctx.Done():
			// context was cancelled during the wait
			logger.Warnf("registration attempt cancelled for service '%s' instance '%s' due to context cancellation: %v", serviceName, instanceId, ctx.Err())
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
	logger.Infof("attempting to deregister service")

	if err := registry.Deregister(shutdownContext, serviceName, instanceId); err != nil {
		logger.Errorf("failed to deregister service '%s' instance '%s': %v", serviceName, instanceId, err)
	} else {
		logger.Infof("successfully deregistered service '%s' instance '%s'", serviceName, instanceId)
	}
}

func setupHLSFolders(cfg cfg.Transcode) {
	if err := os.MkdirAll(cfg.PublicHLSPath, 0777); err != nil {
		logger.Panicf("failed to create public hls folder: %s", err)
	}

	if err := os.MkdirAll(cfg.PrivateHLSPath, 0777); err != nil {
		logger.Panicf("failed to create private hls folder: %s", err)
	}
}
