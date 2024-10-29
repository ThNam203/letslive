package main

import (
	"context"
	"net"
	"os"
	"time"

	// TODO: add swagger
	//_ "sen1or/lets-live/auth/docs"

	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/discovery"
	logger "sen1or/lets-live/auth/logger"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init()

	ctx := context.Background()

	// for consul service discovery
	StartDiscovery(ctx, config.REGISTRY_ADDR)

	dbConn := ConnectDB(ctx)
	defer dbConn.Close(ctx)

	serverAddr := net.JoinHostPort(config.AUTH_SERVER_HOST, config.AUTH_SERVER_PORT)

	server := NewAPIServer(dbConn, serverAddr)
	server.ListenAndServe(false)
	select {}
}

func ConnectDB(ctx context.Context) *pgx.Conn {
	dbConn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		logger.Panicf("unable to connect to database: %v\n", "err", err)
	}

	return dbConn
}

func StartDiscovery(ctx context.Context, serverAddr string) {
	registry, err := discovery.NewConsulRegistry(serverAddr)
	if err != nil {
		logger.Panicf("failed to start discovery mechanism: %s", err)
	}

	serviceName := config.SERVICE_NAME

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, serverAddr, serviceName, instanceID); err != nil {
		logger.Panicf("failed to register server: %s", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := registry.ReportHealthyState(serviceName, instanceID); err != nil {
					logger.Errorf("failed to report healthy state: %s", err)
				}
				time.Sleep(5 * time.Second)
			}
		}
	}()

	<-ctx.Done()

	if err := registry.Deregister(ctx, serviceName, instanceID); err != nil {
		logger.Errorf("failed to deregister service: %s", err)
	}

	cancel()
}
