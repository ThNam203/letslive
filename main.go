package main

import (
	"log"
	"sen1or/lets-live/internal/config"
	loadbalancer "sen1or/lets-live/internal/load-balancer"
	rtmpserver "sen1or/lets-live/internal/rtmp-server"
	webserver "sen1or/lets-live/internal/web-server"
	"sen1or/lets-live/server/api"
	"sen1or/lets-live/server/hack"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var configuration = config.GetConfig()

func main() {
	resetWorkingSpace()

	dbConn := GetDatabaseConnection()
	hack.AutoMigrateAllTables(*dbConn)
	api := api.NewApi(*dbConn)
	go api.ListenAndServe()

	baseDirectory := configuration.PublicHLSPath
	webServerListenAddr := "localhost:" + strconv.Itoa(configuration.WebServerPort)
	allowedSuffixes := [2]string{".ts", ".m3u8"}

	MyWebServer := webserver.NewWebServer(webServerListenAddr, allowedSuffixes[:], baseDirectory)
	go MyWebServer.ListenAndServe()
	go rtmpserver.StartRTMPService()
	go setupTCPLoadBalancer()
	select {}
}

func setupTCPLoadBalancer() {
	lb := loadbalancer.NewTCPLoadBalancer(configuration.LoadBalancer.TCP)
	err := lb.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func GetDatabaseConnection() *gorm.DB {
	var dsn = "host=localhost user=postgres password=postgres dbname=lets-live port=5432 sslmode=disable TimeZone=Asia/Saigon"
	var db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	return db
}
