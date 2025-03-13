package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sen1or/letslive/auth/pkg/discovery"
	"sen1or/letslive/auth/pkg/logger"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	JWT          `yaml:"jwt"`
	Database     `yaml:"database"`
	Verification `yaml:"verification"`
}

func RetrieveConfig(registry discovery.Registry) *Config {
	config, err := retrieveServiceConfig(registry)
	if err != nil {
		logger.Panicf("failed to get config from config server: %s", err)
	}

	config.Database.ConnectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", os.Getenv("AUTH_DB_USER"), os.Getenv("AUTH_DB_PASSWORD"), config.Database.Host, config.Database.Port, config.Database.Name, strings.Join(config.Database.Params, "&"))

	return config
}

// TODO: retry when getting config
func retrieveServiceConfig(registry discovery.Registry) (*Config, error) {
	configserverURL, err := registry.ServiceAddress(context.Background(), "configserver")
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(
		"%s/%s-%s.yml",
		configserverURL,
		"auth_service",
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
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	return &config, nil
}
