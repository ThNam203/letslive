package main

import (
	"log"
	"os"
	webserver "sen1or/lets-live/web-server"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Can not load environment variables")
	}

    baseDirectory := os.Getenv("WEB_SERVER_BASE_DIRECTORY")
	webServerListenAddr := "localhost:" + os.Getenv("WEB_SERVER_PORT")
	allowedSuffixes := [2]string{".ts", ".m3u8"}

	MyWebServer := webserver.NewWebServer(webServerListenAddr, allowedSuffixes[:], baseDirectory)
	MyWebServer.ListenAndServe()
}