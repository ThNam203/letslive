package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/realtime/api"
	"sen1or/letslive/realtime/auth"
	cfg "sen1or/letslive/realtime/config"
	"sen1or/letslive/realtime/hub"
	"sen1or/letslive/realtime/presence"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/natsconn"
	"sen1or/letslive/shared/pkg/realtime"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"
)

var (
	configServiceName = "realtime_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	shutdownTimeout     = 15 * time.Second
	presenceGracePeriod = 5 * time.Second
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
	config := cfgManager.GetConfig()

	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go sharedutils.RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config.Service.Hostname, config.Service.APIPort)

	otelShutdownFunc, err := tracer.SetupOTelSDK(ctx, *config)
	if err != nil {
		logger.Panicf(ctx, "failed to setup otel sdk: %v", err)
	}

	natsConn, err := natsconn.Connect(ctx, config.NATS.URL)
	if err != nil {
		logger.Panicf(ctx, "failed to connect to nats: %v", err)
	}

	publisher := realtime.NewNATSPublisher(natsConn)
	server := api.NewServer(
		config,
		auth.NewVerifier(config.JWKSURL),
		hub.New(hub.NewNATSSource(natsConn), publisher),
		presence.NewTracker(publisher, presenceGracePeriod),
		natsConn.IsConnected,
	)

	go func() {
		logger.Infof(ctx, "starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		if err := server.ListenAndServe(); err != nil {
			logger.Errorf(ctx, "server stopped: %v", err)
		}
		stop()
	}()

	<-ctx.Done()
	logger.Infof(ctx, "shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	var shutdownWg sync.WaitGroup
	shutdownWg.Go(func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Errorf(shutdownCtx, "server shutdown: %v", err)
		}
	})
	shutdownWg.Go(func() {
		sharedutils.DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
	})
	shutdownWg.Go(func() {
		if err := otelShutdownFunc(shutdownCtx); err != nil {
			logger.Errorf(shutdownCtx, "otel shutdown: %v", err)
		}
	})
	shutdownWg.Go(func() {
		if err := natsConn.Drain(); err != nil {
			logger.Errorf(shutdownCtx, "failed to drain nats connection: %v", err)
		}
	})
	shutdownWg.Wait()

	logger.Infof(shutdownCtx, "service shut down complete.")
}
