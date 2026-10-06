package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/chat/api"
	cfg "sen1or/letslive/chat/config"
	"sen1or/letslive/chat/events"
	"sen1or/letslive/chat/gateway/userservice"
	"sen1or/letslive/chat/handlers/conversation"
	"sen1or/letslive/chat/handlers/dmmessage"
	"sen1or/letslive/chat/handlers/general"
	"sen1or/letslive/chat/mongodb"
	"sen1or/letslive/chat/presence"
	"sen1or/letslive/chat/repositories"
	"sen1or/letslive/chat/services"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/natsconn"
	"sen1or/letslive/shared/pkg/realtime"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"
)

var (
	configServiceName = "chat_service"
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

	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go sharedutils.RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config.Service.Hostname, config.Service.APIPort)

	otelShutdownFunc, err := tracer.SetupOTelSDK(ctx, *config)
	if err != nil {
		logger.Panicf(ctx, "failed to setup otel sdk: %v", err)
	}

	mongoClient, err := mongodb.Connect(ctx, config.Database.ConnectionString)
	if err != nil {
		logger.Panicf(ctx, "failed to connect to mongo: %v", err)
	}
	db := mongoClient.Database(config.Database.Name)
	repositories.EnsureIndexes(ctx, db)

	natsConn, err := natsconn.Connect(ctx, config.NATS.URL)
	if err != nil {
		logger.Panicf(ctx, "failed to connect to nats: %v", err)
	}

	conversationRepo := repositories.NewConversationRepository(db)
	dmMessageRepo := repositories.NewDmMessageRepository(db)
	notifier := events.NewNotifier(realtime.NewNATSPublisher(natsConn))

	presenceSub, err := presence.NewSubscriber(conversationRepo, notifier).Subscribe(natsConn)
	if err != nil {
		logger.Panicf(ctx, "failed to subscribe to presence events: %v", err)
	}

	conversationService := services.NewConversationService(conversationRepo, dmMessageRepo, userservice.NewGateway(registry))
	dmMessageService := services.NewDmMessageService(conversationRepo, dmMessageRepo, notifier)

	server := api.NewAPIServer(
		config,
		general.NewGeneralHandler(mongoClient, natsConn.IsConnected),
		conversation.NewConversationHandler(conversationService),
		dmmessage.NewDmMessageHandler(dmMessageService),
	)

	go func() {
		logger.Infof(ctx, "starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		if err := server.ListenAndServe(ctx); err != nil {
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
	shutdownWg.Wait()

	// after the server stops, so in-flight requests can still publish
	if err := presenceSub.Unsubscribe(); err != nil {
		logger.Errorf(shutdownCtx, "failed to unsubscribe from presence events: %v", err)
	}
	if err := natsConn.Drain(); err != nil {
		logger.Errorf(shutdownCtx, "failed to drain nats connection: %v", err)
	}
	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		logger.Errorf(shutdownCtx, "failed to disconnect mongo: %v", err)
	}

	logger.Infof(shutdownCtx, "service shut down complete.")
}
