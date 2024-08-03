package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PublicHLSPath string        `yaml:"publicHLSPath"`
	WebServerPort string        `yaml:"webServerPort"`
	FFMpegSetting FFMpegSetting `yaml:"ffmpegSetting"`
}

type FFMpegSetting struct {
	FFMpegPath     string          `yaml:"ffmpegPath"`
	MasterFileName string          `yaml:"masterFileName"`
	HlsTime        int             `yaml:"hlsTime"`
	Crf            int             `yaml:"crf"`
	Preset         string          `yaml:"preset"`
	HlsListSize    int             `yaml:"hlsListSize"`
	HlsMaxSize     int             `yaml:"hlsMaxSize"`
	Qualities      []FFMpegQuality `yaml:"qualities"`
}

type FFMpegQuality struct {
	Resolution string `yaml:"resolution"`
	MaxBitrate string `yaml:"maxBitrate"`
	FPS        int    `yaml:"fps"`
	BufSize    string `yaml:"bufSize"`
}

func getConfig() Config {
	configurationPath := "config/configuration.yaml"

	if !isFileExists(configurationPath) {
		log.Fatal("configuration.yaml is required")
	}

	configFile, err := os.ReadFile(configurationPath)
	if err != nil {
		panic("error reading configuration file")
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}

	return config
}

// TODO: check config
