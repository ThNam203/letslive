package main

import (
	"log"
	"os"
	"time"
)

func resetWorkingSpace(fileDestination string) {

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
