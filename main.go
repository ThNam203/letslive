package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sen1or/lets-live/internal/config"
	loadbalancer "sen1or/lets-live/internal/load-balancer"
	rtmpserver "sen1or/lets-live/internal/rtmp-server"
	"sen1or/lets-live/server/api"
	"sen1or/lets-live/server/hack"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var cfg = config.GetConfig()

func main() {
	resetWorkingSpace()

	dbConn := GetDatabaseConnection()
	hack.AutoMigrateAllTables(*dbConn)
	api := api.NewApi(*dbConn)
	api.ListenAndServeTLS()

	//webServerListenAddr := "localhost:" + strconv.Itoa(cfg.WebServerPort)
	//allowedSuffixes := [2]string{".ts", ".m3u8"}
	//MyWebServer := webserver.NewWebServer(webServerListenAddr, allowedSuffixes[:], cfg.PrivateHLSPath)
	//MyWebServer.ListenAndServe()

	//ipfsStorage := ipfs.NewIPFSStorage(cfg.PrivateHLSPath, cfg.IPFS.Gateway)
	rtmpserver.StartRTMPService()
	setupTCPLoadBalancer()
	//go internal.MonitorHLSStreamContent(cfg.PrivateHLSPath, ipfsStorage)

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
	lb := loadbalancer.NewTCPLoadBalancer(cfg.LoadBalancer.TCP)
	go (func() {
		err := lb.ListenAndServe()
		if err != nil {
			log.Panic(err)
		}
	})()
}

func GetDatabaseConnection() *gorm.DB {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")

	var dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Saigon", host, user, password, dbname, port)

	var db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	return db
}
