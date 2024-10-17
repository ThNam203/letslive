package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PublicHLSPath  string        `yaml:"publicHLSPath"`
	PrivateHLSPath string        `yaml:"privateHLSPath"`
	WebServerPort  int           `yaml:"webServerPort"`
	ServerURL      string        `yaml:"serverURL"`
	FFMpegSetting  FFMpegSetting `yaml:"ffmpegSetting"`
	IPFS           IPFSSetting   `yaml:"ipfs"`
	LoadBalancer   LoadBalancer  `yaml:"loadBalancer"`
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

type IPFSSetting struct {
	Enabled           bool   `yaml:"enabled"`
	Gateway           string `yaml:"gateway"`
	BootstrapNodeAddr string `yaml:"bootstrapNodeAddr"`
}

type LoadBalancer struct {
	HTTP LBSetting `yaml:"http"`
	TCP  LBSetting `yaml:"tcp"`
}

type LBSetting struct {
	Name string `yaml:"name"`
	From string `yaml:"from"`
	To   []string
}

var configuration *Config = nil

func GetConfig() Config {
	if configuration != nil {
		return *configuration
	}
	configurationPath := "core/config/configuration.yaml"

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
	if config.PublicHLSPath == "" {
		log.Fatal("empty PublicHLSPath")
	}

	if config.PrivateHLSPath == "" {
		log.Fatal("empty PrivateHLSPath")
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
