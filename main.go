package main

import (
	config "sen1or/lets-live/config"
	rtmpserver "sen1or/lets-live/rtmp-server"
	webserver "sen1or/lets-live/web-server"
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

	rtmpserver.StartRTMPService()
}
