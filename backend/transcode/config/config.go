package config

import (
	"fmt"
	"io"
	"net/http"
	"sen1or/lets-live/pkg/logger"

	"gopkg.in/yaml.v3"
)

const (
	CONFIG_SERVER_PROTOCOL             = "http"
	CONFIG_SERVER_ADDRESS              = "configserver:8181"
	CONFIG_SERVER_SERVICE_APPLICATION  = "transcode_service"
	CONFIG_SERVER_SERVICE_PROFILE      = "default"
	CONFIG_SERVER_REGISTRY_APPLICATION = "registry_service"
	CONFIG_SERVER_REGISTRY_PROFILE     = "default"
)

type RegistryConfig struct {
	Service struct {
		Address string   `yaml:"address"`
		Tags    []string `yaml:"tags"`
	} `yaml:"registry"`
}

type Config struct {
	Service struct {
		Name            string `yaml:"name"`
		Hostname        string `yaml:"hostname"`
		APIPort         int    `yaml:"apiPort"`
		RtmpBindAddress string `yaml:"rtmpBindAddress"`
		Port            int    `yaml:"port"`
	} `yaml:"service"`
	Registry RegistryConfig
	RTMP     struct {
		Port               int    `yaml:"port"`
		UserServiceAddress string `yaml:"userServiceAddress"`
	} `yaml:"rtmp"`
	Transcode struct {
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
	} `yaml:"transcode"`
	IPFS struct {
		Enabled           bool   `yaml:"enabled"`
		Gateway           string `yaml:"gateway"` // the gateway address, it is used to generate the final url to the ipfs file
		BootstrapNodeAddr string `yaml:"bootstrapNodeAddr"`
	} `yaml:"ipfs"`
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
	url := fmt.Sprintf("%s://%s/%s-%s.yml", CONFIG_SERVER_PROTOCOL, CONFIG_SERVER_ADDRESS, CONFIG_SERVER_SERVICE_APPLICATION, CONFIG_SERVER_SERVICE_PROFILE)

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

	// -------------------
	registryConfig, err := retrieveRegistryConfig()
	if err != nil {
		return nil, err
	}

	config.Registry = *registryConfig

	return &config, nil
}

func retrieveRegistryConfig() (*RegistryConfig, error) {
	url := fmt.Sprintf("%s://%s/%s-%s.yml", CONFIG_SERVER_PROTOCOL, CONFIG_SERVER_ADDRESS, CONFIG_SERVER_REGISTRY_APPLICATION, CONFIG_SERVER_REGISTRY_PROFILE)

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
