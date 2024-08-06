package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PublicHLSPath string        `yaml:"publicHLSPath"`
	WebServerPort int           `yaml:"webServerPort"`
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

var configuration *Config = nil

func GetConfig() Config {
	if configuration != nil {
		return *configuration
	}
	configurationPath := "config/configuration.yaml"

	if !isFileExists(configurationPath) {
		log.Fatal("configuration.yaml is required")
	}

	configFile, err := os.ReadFile(configurationPath)
	if err != nil {
		log.Fatal("error reading configuration file")
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)

	if err != nil {
		log.Fatal(err)
	}

	checkConfig(config)
	configuration = &config

	return config
}

func checkConfig(config Config) {
	if !isFileExists(config.PublicHLSPath) {
		log.Fatalf("PublicHLSPath %s doesn't exist", config.PublicHLSPath)
	}

	if !isFileExists(config.FFMpegSetting.FFMpegPath) {
		log.Fatalf("ffmpeg path %s doesn't exist", config.FFMpegSetting.FFMpegPath)
	}
}

func isFileExists(path string) bool {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		log.Println(path)
		return false
	}

	return true
}
