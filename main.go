package main

import (
	"log"
	"sen1or/lets-live/internal/config"
	loadbalancer "sen1or/lets-live/internal/load-balancer"
	rtmpserver "sen1or/lets-live/internal/rtmp-server"
	webserver "sen1or/lets-live/internal/web-server"
	"strconv"
)

var configuration = config.GetConfig()

func main() {
	resetWorkingSpace()

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
