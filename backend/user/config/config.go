package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sen1or/letslive/user/pkg/discovery"
	"sen1or/letslive/user/pkg/logger"
	"strings"

	"gopkg.in/yaml.v3"
)

type MinIO struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	ReturnURL string `yaml:"returnURL"`
}

type Service struct {
	Name           string `yaml:"name"`
	Hostname       string `yaml:"hostname"`
	APIBindAddress string `yaml:"apiBindAddress"`
	APIPort        int    `yaml:"apiPort"`
}

type Database struct {
	MigrationPath    string   `yaml:"migration-path"`
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Name             string   `yaml:"name"`
	Params           []string `yaml:"params"`
	ConnectionString string
}

type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	MinIO    `yaml:"minio"`
}

func RetrieveConfig(registry discovery.Registry) *Config {
	config, err := retrieveConfig(registry)
	if err != nil {
		logger.Panicf("failed to get config from config server: %s", err)
	}

	config.Database.ConnectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", os.Getenv("USER_DB_USER"), os.Getenv("USER_DB_PASSWORD"), config.Database.Host, config.Database.Port, config.Database.Name, strings.Join(config.Database.Params, "&"))

	return config
}

func retrieveConfig(registry discovery.Registry) (*Config, error) {
	configserverURL, err := registry.ServiceAddress(context.Background(), "configserver")
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(
		"%s/%s-%s.yml",
		configserverURL,
		"user_service",
		os.Getenv("CONFIG_SERVER_PROFILE"),
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

	return &config, nil
}
