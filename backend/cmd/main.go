package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	httpServer "sen1or/lets-live/api/http"
	"sen1or/lets-live/core/config"
	"sen1or/lets-live/core/rtmp"
	"sen1or/lets-live/core/storage/ipfs"
	"sen1or/lets-live/core/watcher"

	loadbalancer "sen1or/lets-live/core/load-balancer"
	"sen1or/lets-live/core/logger"
	webserver "sen1or/lets-live/core/web-server"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var cfg = config.GetConfig()

func main() {
	logger.Init()
	resetWorkingSpace()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go RunBackend()

	allowedSuffixes := [2]string{".ts", ".m3u8"}
	MyWebServer := webserver.NewWebServer(cfg.WebServerPort, allowedSuffixes[:], cfg.PublicHLSPath)
	MyWebServer.ListenAndServe()

	//ipfsStorage := ipfs.NewKuboStorage(cfg.PrivateHLSPath, cfg.IPFS.Gateway)
	ipfsStorage := ipfs.NewCustomStorage(ctx, cfg.IPFS.Gateway, cfg.IPFS.BootstrapNodeAddr)
	monitor := watcher.NewStreamWatcher(cfg.PrivateHLSPath, ipfsStorage)
	go monitor.MonitorHLSStreamContent()

	setupTCPLoadBalancer()
	rtmpServer := rtmp.NewRTMPServer(1936, cfg.ServerURL)
	go rtmpServer.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			os.Exit(1)
		}
	}()

	log.Println("Setup all done!")
	select {}
}

func setupTCPLoadBalancer() {
	go (func() {
		lb := loadbalancer.NewTCPLoadBalancer(cfg.LoadBalancer.TCP)

		err := lb.ListenAndServe()
		if err != nil {
			log.Panic(err)
		}
	})()
}

func RunBackend() {
	dbConn := GetDatabaseConnection()
	hack.AutoMigrateAllTables(*dbConn)
	server := httpServer.NewAPIServer(*dbConn)
	server.ListenAndServeTLS()
}

func GetDatabaseConnection() *gorm.DB {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")

	var dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Saigon", host, user, password, dbname, port)

	var db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DryRun: false,
	})
	if err != nil {
		log.Panic(err)
	}

	return db
}
