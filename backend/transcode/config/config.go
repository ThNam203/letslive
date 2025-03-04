package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sen1or/lets-live/pkg/logger"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Name            string `yaml:"name"`
	Hostname        string `yaml:"hostname"`
	APIPort         int    `yaml:"apiPort"`
	RtmpBindAddress string `yaml:"rtmpBindAddress"`
	Port            int    `yaml:"port"`
}

type RTMP struct {
	Port               int    `yaml:"port"`
	UserServiceAddress string `yaml:"userServiceAddress"`
}

type RegistryConfig struct {
	Service struct {
		Address string   `yaml:"address"`
		Tags    []string `yaml:"tags"`
	} `yaml:"registry"`
}

type IPFS struct {
	Enabled           bool     `yaml:"enabled"`
	Gateway           string   `yaml:"gateway"` // the gateway address, it is used to generate the final url to the ipfs file
	SubGateways       []string `yaml:"subGateways"`
	BootstrapNodeAddr string   `yaml:"bootstrapNodeAddr"`
}

type MinIO struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ClientHost string `yaml:"clientHost"` // it is used for development only, cause minio:9090 (Host) and ui outside of docker can't get files, in real scenario it should be equal to Host
	BucketName string `yaml:"bucketName"`
	AccessKey  string `yaml:"accessKey"`
	SecretKey  string `yaml:"secretKey"`
}

type Transcode struct {
	PublicHLSPath  string `yaml:"publicHLSPath"`
	PrivateHLSPath string `yaml:"privateHLSPath"`
	FFMpegSetting  struct {
		FFMpegPath     string `yaml:"ffmpegPath"`
		MasterFileName string `yaml:"masterFileName"`
		HLSTime        int    `yaml:"hlsTime"`
		CRF            int    `yaml:"crf"`
		Preset         string `yaml:"preset"`
		HlsListSize    int    `yaml:"hlsListSize"`
		HlsMaxSize     int    `yaml:"hlsMaxSize"`
		Qualities      []struct {
			Resolution string `yaml:"resolution"`
			MaxBitrate string `yaml:"maxBitrate"`
			FPS        int    `yaml:"fps"`
			BufSize    string `yaml:"bufSize"`
		} `yaml:"qualities"`
	} `yaml:"ffmpegSetting"`
}

type Config struct {
	Service   `yaml:"service"`
	Registry  RegistryConfig
	RTMP      `yaml:"rtmp"`
	Transcode `yaml:"transcode"`
	IPFS      `yaml:"ipfs"`
	MinIO     `yaml:"minio"`
	Webserver struct {
		Port int `yaml:"port"`
	} `yaml:"webserver"`
}

func RetrieveConfig() *Config {
	config, err := retrieveConfig()
	if err != nil {
		logger.Panicf("failed to get config: %s", err)
	}

	return config
}

func retrieveConfig() (*Config, error) {
	url := fmt.Sprintf(
		"%s://%s/%s-%s.yml",
		os.Getenv("CONFIG_SERVER_PROTOCOL"),
		os.Getenv("CONFIG_SERVER_ADDRESS"),
		os.Getenv("CONFIG_SERVER_SERVICE_APPLICATION"),
		os.Getenv("CONFIG_SERVER_SERVICE_PROFILE"),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to config server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	registryConfig, err := retrieveRegistryConfig()
	if err != nil {
		return nil, err
	}

	config.Registry = *registryConfig

	return &config, nil
}

func retrieveRegistryConfig() (*RegistryConfig, error) {
	url := fmt.Sprintf(
		"%s://%s/%s-%s.yml",
		os.Getenv("CONFIG_SERVER_PROTOCOL"),
		os.Getenv("CONFIG_SERVER_ADDRESS"),
		os.Getenv("CONFIG_SERVER_REGISTRY_APPLICATION"),
		os.Getenv("CONFIG_SERVER_REGISTRY_PROFILE"),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("error while creating request: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to config server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var config RegistryConfig
	err = yaml.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return &config, nil
}
