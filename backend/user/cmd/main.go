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
	"sen1or/letslive/user/domains"
	financehttp "sen1or/letslive/user/gateway/finance/http"
	devicehandler "sen1or/letslive/user/handlers/device"
	"sen1or/letslive/user/handlers/follow"
	gifthandler "sen1or/letslive/user/handlers/gift"
	inventoryhandler "sen1or/letslive/user/handlers/inventory"
	"sen1or/letslive/user/handlers/livestream_information"
	notificationhandler "sen1or/letslive/user/handlers/notification"
	pushhandler "sen1or/letslive/user/handlers/push"
	"sen1or/letslive/user/handlers/user"
	"sen1or/letslive/user/push"
	"sen1or/letslive/user/repositories"
	"sen1or/letslive/user/services"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/natsconn"
	"sen1or/letslive/shared/pkg/realtime"
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

	natsConn, err := natsconn.Connect(ctx, config.NATS.URL)
	if err != nil {
		logger.Panicf(ctx, "failed to connect to nats: %v", err)
	}
	realtimePublisher := realtime.NewNATSPublisher(natsConn)

	deviceRepo := repositories.NewDeviceTokenRepository(dbConn)
	pushSender, closePush, err := push.NewFromEnv(ctx, deviceRepo, natsConn)
	if err != nil {
		logger.Panicf(ctx, "invalid push configuration: %v", err)
	}

	server := SetupServer(ctx, dbConn, registry, config, realtimePublisher, deviceRepo, pushSender)
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
		// After Shutdown returns (or times out): requests that finished have
		// enqueued their pushes; let the queue deliver them.
		closePush()
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
	// nats is drained last so queued pushes can still read presence
	shutdownWg.Add(1)
	go (func() {
		if err := natsConn.Drain(); err != nil {
			logger.Errorf(shutdownCtx, "failed to drain nats connection: %v", err)
		}
		shutdownWg.Done()
	})()

	shutdownWg.Wait()
	logger.Infof(shutdownCtx, "service shut down complete.")
}

func SetupServer(ctx context.Context, dbConn *pgxpool.Pool, registry discovery.Registry, cfg *cfg.Config, publisher realtime.Publisher, deviceRepo domains.DeviceTokenRepository, pushSender push.Sender) *api.APIServer {
	var userRepo = repositories.NewUserRepository(dbConn)
	var livestreamInfoRepo = repositories.NewLivestreamInformationRepository(dbConn)
	var followRepo = repositories.NewFollowRepository(dbConn)
	var notificationRepo = repositories.NewNotificationRepository(dbConn)
	var inventoryRepo = repositories.NewInventoryRepository(dbConn)
	var giftRepo = repositories.NewGiftRepository(dbConn)

	minioService := services.NewMinIOService(ctx, cfg.MinIO)
	var userService = services.NewUserService(userRepo, livestreamInfoRepo, notificationRepo, followRepo, *minioService)
	var livestreamInfoService = services.NewLivestreamInformationService(livestreamInfoRepo)
	var followService = services.NewFollowService(followRepo)
	var notificationService = services.NewNotificationService(notificationRepo, publisher, pushSender)
	var inventoryService = services.NewInventoryService(inventoryRepo)
	var deviceService = services.NewDeviceService(deviceRepo)
	var financeGateway = financehttp.NewFinanceGateway(registry)
	var giftService = services.NewGiftService(giftRepo, userRepo, financeGateway, notificationService)

	var userHandler = user.NewUserHandler(*userService)
	var livestreamInfoHandler = livestream_information.NewLivestreamInformationHandler(*livestreamInfoService, *minioService)
	var followHandler = follow.NewFollowHandler(*followService)
	var notifHandler = notificationhandler.NewNotificationHandler(*notificationService)
	var invHandler = inventoryhandler.NewInventoryHandler(inventoryService)
	var gHandler = gifthandler.NewGiftHandler(giftService)
	var dHandler = devicehandler.NewDeviceHandler(deviceService)
	return api.NewAPIServer(userHandler, livestreamInfoHandler, followHandler, notifHandler, invHandler, gHandler, dHandler, pushhandler.NewPushHandler(pushSender), cfg, dbConn)
}
