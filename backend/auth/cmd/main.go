package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/auth/api"
	cfg "sen1or/letslive/auth/config"
	"sen1or/letslive/auth/handlers"
	"sen1or/letslive/auth/repositories"
	"sen1or/letslive/auth/services"

	usergateway "sen1or/letslive/auth/gateway/user/http"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "auth_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	shutdownTimeout = 15 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Init(logger.LogLevel(logger.Debug))
	validateJWTSecrets(ctx)

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

	server := SetupServer(dbConn, registry, config)
	go func() {
		logger.Infof(ctx, "starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		// ListenAndServe should ideally block until an error occurs (e.g., server stopped)
		server.ListenAndServe(ctx, false)
		stop() // trigger shutdown if server fails unexpectedly
	}()

	logger.Infof(ctx, "server started.")
	<-ctx.Done() // block here until SIGINT/SIGTERM is received (ctx from signal.NotifyContext)
	// initiate graceful shutdown
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

func validateJWTSecrets(ctx context.Context) {
	secrets := map[string]string{
		"ACCESS_TOKEN_SECRET":       os.Getenv("ACCESS_TOKEN_SECRET"),
		"REFRESH_TOKEN_SECRET":      os.Getenv("REFRESH_TOKEN_SECRET"),
		"REACTIVATION_TOKEN_SECRET": os.Getenv("REACTIVATION_TOKEN_SECRET"),
	}

	for name, value := range secrets {
		if value == "" {
			logger.Panicf(ctx, "missing required env var: %s must be set", name)
		}
	}

	if secrets["ACCESS_TOKEN_SECRET"] == secrets["REFRESH_TOKEN_SECRET"] {
		logger.Panicf(ctx, "ACCESS_TOKEN_SECRET and REFRESH_TOKEN_SECRET must be distinct")
	}
	if secrets["ACCESS_TOKEN_SECRET"] == secrets["REACTIVATION_TOKEN_SECRET"] {
		logger.Panicf(ctx, "ACCESS_TOKEN_SECRET and REACTIVATION_TOKEN_SECRET must be distinct")
	}
	if secrets["REFRESH_TOKEN_SECRET"] == secrets["REACTIVATION_TOKEN_SECRET"] {
		logger.Panicf(ctx, "REFRESH_TOKEN_SECRET and REACTIVATION_TOKEN_SECRET must be distinct")
	}
}

func SetupServer(dbConn *pgxpool.Pool, registry discovery.Registry, cfg *cfg.Config) *api.APIServer {
	var userRepo = repositories.NewAuthRepository(dbConn)
	var refreshTokenRepo = repositories.NewRefreshTokenRepository(dbConn)
	var signUpOTPRepo = repositories.NewSignUpOTPRepo(dbConn)

	userGateway := usergateway.NewUserGateway(registry)
	var authService = services.NewAuthService(userRepo, userGateway)
	var googleAuthService = services.NewGoogleAuthService(userRepo, userGateway)
	var jwtService = services.NewJWTService(refreshTokenRepo, cfg.JWT, userGateway)
	var verificationService = services.NewVerificationService(signUpOTPRepo)
	var authHandler = handlers.NewAuthHandler(*jwtService, *authService, *verificationService, *googleAuthService, cfg.Verification.Gateway)
	return api.NewAPIServer(authHandler, registry, cfg, dbConn)
}
