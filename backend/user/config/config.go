package config

import (
	"fmt"
	"io"
	"net/http"
	"sen1or/lets-live/pkg/logger"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	CONFIG_SERVER_PROTOCOL             = "http"
	CONFIG_SERVER_ADDRESS              = "configserver:8181"
	CONFIG_SERVER_APPLICATION          = "user_service"
	CONFIG_SERVER_PROFILE              = "default"
	CONFIG_SERVER_REGISTRY_APPLICATION = "registry_service"
	CONFIG_SERVER_REGISTRY_PROFILE     = "default"
)

type RegistryConfig struct {
	RegistryService struct {
		Address string   `yaml:"address"`
		Tags    []string `yaml:"tags"`
	} `yaml:"registry"`
}

type MinIO struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ClientHost string `yaml:"clientHost"` // it is used for development only, cause minio:9090 (Host) and ui outside of docker can't get files, in real scenario it should be equal to Host
	AccessKey  string `yaml:"accessKey"`
	SecretKey  string `yaml:"secretKey"`
}

type Service struct {
	Name           string `yaml:"name"`
	Hostname       string `yaml:"hostname"`
	APIBindAddress string `yaml:"apiBindAddress"`
	APIPort        int    `yaml:"apiPort"`
}

type Database struct {
	MigrationPath    string   `yaml:"migration-path"`
	User             string   `yaml:"user"`
	Password         string   `yaml:"password"`
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Name             string   `yaml:"name"`
	Params           []string `yaml:"params"`
	ConnectionString string
}

type Config struct {
	Service  `yaml:"service"`
	Registry RegistryConfig
	Database `yaml:"database"`
	MinIO    `yaml:"minio"`
}

func RetrieveConfig() *Config {
	config, err := retrieveConfig()
	if err != nil {
		logger.Panicf("failed to get config from config server: %s", err)
	}

	config.Database.ConnectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name, strings.Join(config.Database.Params, "&"))

	registryConfig, err := retrieveRegistryConfig()
	if err != nil {
		logger.Panicf("failed to get registry config: %s", err)
	}
	config.Registry = *registryConfig

	return config
}

func retrieveConfig() (*Config, error) {
	url := fmt.Sprintf("%s://%s/%s-%s.yml", CONFIG_SERVER_PROTOCOL, CONFIG_SERVER_ADDRESS, CONFIG_SERVER_APPLICATION, CONFIG_SERVER_PROFILE)

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
