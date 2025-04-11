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

	shutdownTimeout            = 15 * time.Second
	discoveryDeregisterTimeout = 10 * time.Second
	gracefulShutdownTimeout    = 15 * time.Second
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
		rtmp.RTMPServerConfig{Port: config.RTMP.Port, Registry: &registry, Config: *config, VODHandler: vodHandler},
		userGateway,
		livestreamGateway,
	)
	go rtmpServer.Start()
	<-ctx.Done()

	logger.Infof("starting coordinated shutdown...")

	// Create a shutdown context with timeout
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), gracefulShutdownTimeout) // Adjust timeout
	defer cancelShutdown()

	// Use a WaitGroup if shutting down components in parallel
	var wg sync.WaitGroup // Import "sync"

	// Shutdown Web Server (if it has a Shutdown method)
	wg.Add(1)
	go func() {
		defer wg.Done()
		MyWebServer.Shutdown(shutdownCtx)
	}()

	// Shutdown RTMP Server (if it has a Shutdown/Stop method)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rtmpServer.Shutdown(shutdownCtx)
	}()

	// Stop Watcher (if it has a Shutdown/Stop method)
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Shutdown()
	}()

	wg.Wait()

	logger.Infof("service shutdown complete.")
}

func RegisterToDiscoveryService(ctx context.Context, registry discovery.Registry, config *cfg.Config) {
	serviceName := config.Service.Name
	serviceHostPort := fmt.Sprintf("%s:%d", config.Service.Hostname, config.Webserver.Port)
	serviceHealthCheckURL := fmt.Sprintf("http://%s/v1/health", serviceHostPort)
	instanceID := discovery.GenerateInstanceID(config.Service.Name)

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
	logger.Infof("service '%s' instance '%s' registered - waiting for context cancellation signal to deregister...", serviceName, instanceID)
	<-ctx.Done() // Wait for shutdown signal

	logger.Infof("context cancelled - attempting to deregister service '%s' instance '%s'...", serviceName, instanceID)

	// Create a new short-lived context for the deregistration call.
	// The original context `ctx` is already cancelled, so using it might cause immediate failure.
	// Use context.Background() as the parent to ensure it's not tied to the cancelled context.
	deregisterCtx, cancelDeregister := context.WithTimeout(context.Background(), discoveryDeregisterTimeout)
	defer cancelDeregister()

	if err := registry.Deregister(deregisterCtx, serviceName, instanceID); err != nil {
		logger.Errorf("failed to deregister service '%s' instance '%s': %v", serviceName, instanceID, err)
	} else {
		logger.Infof("successfully deregistered service '%s' instance '%s'", serviceName, instanceID)
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
