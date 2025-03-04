package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sen1or/lets-live/pkg/logger"
	"strings"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	Address string `yaml:"address"`
}

type JWT struct {
	RefreshTokenMaxAge int    `yaml:"refresh-token-max-age"`
	AccessTokenMaxAge  int    `yaml:"access-token-max-age"`
	Consumer           string `yaml:"consumer"`
	Issuer             string `yaml:"issuer"`
	Subject            string `yaml:"subject"`
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

type Verification struct {
	Gateway string `yaml:"gateway"`
}

type Config struct {
	Service      `yaml:"service"`
	Registry     `yaml:"registry"`
	JWT          `yaml:"jwt"`
	Database     `yaml:"database"`
	Verification `yaml:"verification"`
}

func RetrieveConfig() *Config {
	config, err := retrieveServiceConfig()
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

func retrieveServiceConfig() (*Config, error) {
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
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	return &config, nil
}

func retrieveRegistryConfig() (*Registry, error) {
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

	var config Config
	err = yaml.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return &config.Registry, nil
}
