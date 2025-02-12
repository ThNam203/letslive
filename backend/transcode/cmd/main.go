package main

import (
	"context"
	"fmt"
	"os"
	"sen1or/lets-live/pkg/logger"
	cfg "sen1or/lets-live/transcode/config"
	usergateway "sen1or/lets-live/transcode/gateway/user/http"
	"sen1or/lets-live/transcode/rtmp"
	ipfsstorage "sen1or/lets-live/transcode/storage/ipfs"
	miniostorage "sen1or/lets-live/transcode/storage/minio"
	"sen1or/lets-live/transcode/watcher"
	ipfswatcher "sen1or/lets-live/transcode/watcher/ipfs"
	miniowatcher "sen1or/lets-live/transcode/watcher/minio"
	"sen1or/lets-live/transcode/webserver"

	"sen1or/lets-live/pkg/discovery"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()

	logger.Init(logger.LogLevel(logger.Debug))
	config := cfg.RetrieveConfig()

	//if err := resetWorkingSpace(*config); err != nil {
	//	logger.Panicf("failed to reset working space: %s", err)
	//}

	registry, err := discovery.NewConsulRegistry(config.Registry.Service.Address)
	if err != nil {
		logger.Panicf("failed to get a new registry")
	}

	// TODO: fix this, we need a webserver built into the transcode (not the pkg/webserver, use nginx instead)
	serviceHostPort := fmt.Sprintf("%s:%d", config.Service.Hostname, config.Webserver.Port)
	serviceHealthCheckURL := fmt.Sprintf("http://%s/v1/health", serviceHostPort)
	instanceID := discovery.GenerateInstanceID(config.Service.Name)
	registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, config.Service.Name, instanceID, config.Registry.Service.Tags)

	allowedSuffixes := [2]string{".ts", ".m3u8"}
	MyWebServer := webserver.NewWebServer(config.Webserver.Port, allowedSuffixes[:], config.Transcode.PublicHLSPath)
	MyWebServer.ListenAndServe()

	// TODO: find a way to remove the ipfsVOD from the rtmp, or change the design or config
	var vodHandler watcher.VODHandler

	if config.IPFS.Enabled {
		vodHandler = ipfswatcher.GetIPFSVODHandler()
		ipfsStorage := ipfsstorage.NewIPFSStorage(context.Background(), config.IPFS.Gateway, &config.IPFS.BootstrapNodeAddr)
		ipfsWatcherStrategy := ipfswatcher.NewIPFSStorageWatcherStrategy(vodHandler, ipfsStorage, *config)
		watcher := watcher.NewFFMpegFileWatcher(config.Transcode.PrivateHLSPath, ipfsWatcherStrategy)
		go watcher.Watch()
	} else {
		vodHandler = miniowatcher.GetMinIOVODStrategy()
		minioStorage := miniostorage.NewMinIOStorage(context.Background())
		minioWatcherStrategy := miniowatcher.NewMinIOFileWatcherStrategy(vodHandler, minioStorage, *config)
		watcher := watcher.NewFFMpegFileWatcher(config.Transcode.PrivateHLSPath, minioWatcherStrategy)
		go watcher.Watch()
	}

	userGateway := usergateway.NewUserGateway(registry)

	// TODO: find a way to remove the ipfsVOD from the rtmp, or change the design or config
	rtmpServer := rtmp.NewRTMPServer(
		rtmp.RTMPServerConfig{Port: config.RTMP.Port, Registry: &registry, Config: *config, VODHandler: vodHandler},
		userGateway,
	)
	go rtmpServer.Start()
	select {}
}

func resetWorkingSpace(config cfg.Config) error {
	if err := os.RemoveAll(config.Transcode.PublicHLSPath); err != nil {
		return err
	}

	if err := os.RemoveAll(config.Transcode.PrivateHLSPath); err != nil {
		return err
	}

	if err := os.MkdirAll(config.Transcode.PublicHLSPath, 0777); err != nil {
		return err
	}

	if err := os.MkdirAll(config.Transcode.PrivateHLSPath, 0777); err != nil {
		return err
	}

	return nil
}
