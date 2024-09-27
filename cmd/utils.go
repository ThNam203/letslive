package main

import (
	"log"
	"os"
	"sen1or/lets-live/internal/config"
	"time"
)

func resetWorkingSpace() {
	var config = config.GetConfig()
	if err := os.RemoveAll(config.PublicHLSPath); err != nil {
		log.Fatal(err)
	}

	if err := os.RemoveAll(config.PrivateHLSPath); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(config.PublicHLSPath, 0777); err != nil {
		log.Panic(err)
	}

	if err := os.MkdirAll(config.PrivateHLSPath, 0777); err != nil {
		log.Panic(err)
	}
}

func touch(fileDestination string) {
	_, err := os.Stat(fileDestination)
	if os.IsNotExist(err) {
		file, err := os.Create(fileDestination)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err := os.Chtimes(fileDestination, currentTime, currentTime)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func isFileExists(path string) bool {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
