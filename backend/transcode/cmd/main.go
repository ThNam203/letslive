package main

import (
	"context"
	"fmt"
	"os"
	cfg "sen1or/letslive/transcode/config"
	livestreamgateway "sen1or/letslive/transcode/gateway/livestream/http"
	usergateway "sen1or/letslive/transcode/gateway/user/http"
	"sen1or/letslive/transcode/pkg/discovery"
	"sen1or/letslive/transcode/pkg/logger"
	"sen1or/letslive/transcode/rtmp"
	ipfsstorage "sen1or/letslive/transcode/storage/ipfs"
	miniostorage "sen1or/letslive/transcode/storage/minio"
	"sen1or/letslive/transcode/watcher"
	ipfswatcher "sen1or/letslive/transcode/watcher/ipfs"
	miniowatcher "sen1or/letslive/transcode/watcher/minio"
	"sen1or/letslive/transcode/webserver"
)

func main() {
	ctx := context.Background()

	logger.Init(logger.LogLevel(logger.Debug))
	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf("failed to get a new registry")
	}

	config := cfg.RetrieveConfig(registry)
	setupHLSFolders(config.Transcode)

	// TODO: fix this, we need a webserver built into the transcode (not the pkg/webserver, use nginx instead)
	serviceHostPort := fmt.Sprintf("%s:%d", config.Service.Hostname, config.Webserver.Port)
	serviceHealthCheckURL := fmt.Sprintf("http://%s/v1/health", serviceHostPort)
	instanceID := discovery.GenerateInstanceID(config.Service.Name)
	registry.Register(ctx, serviceHostPort, serviceHealthCheckURL, config.Service.Name, instanceID, nil)

	allowedSuffixes := [2]string{".ts", ".m3u8"}
	MyWebServer := webserver.NewWebServer(config.Webserver.Port, allowedSuffixes[:], config.Transcode.PublicHLSPath)
	MyWebServer.ListenAndServe()

	var vodHandler watcher.VODHandler

	if config.IPFS.Enabled {
		vodHandler = ipfswatcher.GetIPFSVODHandler(config.IPFS)
		ipfsStorage := ipfsstorage.NewIPFSStorage(context.Background(), config.IPFS.Gateway, &config.IPFS.BootstrapNodeAddr)
		ipfsWatcherStrategy := ipfswatcher.NewIPFSStorageWatcherStrategy(vodHandler, ipfsStorage, *config)
		watcher := watcher.NewFFMpegFileWatcher(config.Transcode.PrivateHLSPath, ipfsWatcherStrategy)
		go watcher.Watch()
	} else {
		vodHandler = miniowatcher.GetMinIOVODStrategy()
		minioStorage := miniostorage.NewMinIOStorage(context.Background(), config.MinIO)
		minioWatcherStrategy := miniowatcher.NewMinIOFileWatcherStrategy(vodHandler, minioStorage, *config)
		watcher := watcher.NewFFMpegFileWatcher(config.Transcode.PrivateHLSPath, minioWatcherStrategy)
		go watcher.Watch()
	}

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
	select {}
}

func setupHLSFolders(cfg cfg.Transcode) {
	if err := os.MkdirAll(cfg.PublicHLSPath, 0777); err != nil {
		logger.Panicf("failed to create public hls folder: %s", err)
	}

	if err := os.MkdirAll(cfg.PrivateHLSPath, 0777); err != nil {
		logger.Panicf("failed to create private hls folder: %s", err)
	}
}
