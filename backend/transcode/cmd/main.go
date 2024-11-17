package main

import (
	"context"
	"fmt"
	"os"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/pkg/webserver"
	cfg "sen1or/lets-live/transcode/config"
	"sen1or/lets-live/transcode/rtmp"
	"sen1or/lets-live/transcode/storage/ipfs"
	"sen1or/lets-live/transcode/watcher"

	"sen1or/lets-live/pkg/discovery"
)

func main() {
	ctx := context.Background()

	logger.Init()
	config := cfg.RetrieveConfig()

	if err := resetWorkingSpace(*config); err != nil {
		logger.Panicf("failed to reset working space: %s", err)
	}

	registry, err := discovery.NewConsulRegistry(config.Registry.Address)
	if err != nil {
		logger.Panicf("failed to get a new registry")
	}

	serviceHealthCheckURL := fmt.Sprintf("http://%s:%s/v1/health", config.Service.Hostname, config.Service.APIPort)
	instanceID := discovery.GenerateInstanceID(config.Service.Name)
	registry.Register(ctx, config.Registry.Address, serviceHealthCheckURL, config.Service.Name, instanceID)

	allowedSuffixes := [2]string{".ts", ".m3u8"}
	MyWebServer := webserver.NewWebServer(config.Webserver.Port, allowedSuffixes[:], config.Transcode.PublicHLSPath)
	MyWebServer.ListenAndServe()

	//ipfsStorage := ipfs.NewKuboStorage(cfg.PrivateHLSPath, cfg.IPFS.Gateway)
	//ipfsStorage := ipfs.NewCustomStorage(ctx, config.IPFS.Gateway, config.IPFS.BootstrapNodeAddr)
	ipfsStorage := ipfs.NewCustomStorage(ctx, config.IPFS.Gateway, nil)
	monitor := watcher.NewStreamWatcher(config.Transcode.PrivateHLSPath, ipfsStorage, *config)
	go monitor.MonitorHLSStreamContent()

	rtmpServer := rtmp.NewRTMPServer(rtmp.RTMPServerConfig{Port: config.RTMP.Port, Registry: registry, Config: *config})
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
