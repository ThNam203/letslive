package config

import (
	"fmt"
	"io"
	"net/http"
	"sen1or/lets-live/auth/logger"
	"strings"

	"gopkg.in/yaml.v3"
)

var MyConfig *Config

const (
	CONFIG_SERVER_PROTOCOL    = "http"
	CONFIG_SERVER_ADDRESS     = "172.17.0.1:8181"
	CONFIG_SERVER_APPLICATION = "auth_service"
	CONFIG_SERVER_PROFILE     = "default"
)

type Config struct {
	Service struct {
		Name    string `yaml:"name"`
		Address string `yaml:"address"`
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
	} `yaml:"service"`
	Registry struct {
		Address string   `yaml:"address"`
		Tags    []string `yaml:"tags"`
	} `yaml:"registry"`
	Tokens struct {
		RefreshTokenExpiresDuration string `yaml:"refresh-token-expires-duration"`
		AccessTokenExpiresDuration  string `yaml:"access-token-expires-duration"`
		RefreshTokenMaxAge          int    `yaml:"refresh-token-max-age"`
		AccessTokenMaxAge           int    `yaml:"access-token-max-age"`
	} `yaml:"tokens"`
	SSL struct {
		ServerCrtFile string `yaml:"server-crt-file"`
		ServerKeyFile string `yaml:"server-key-file"`
	} `yaml:"ssl"`
	Database struct {
		User             string   `yaml:"user"`
		Password         string   `yaml:"password"`
		Host             string   `yaml:"host"`
		Port             int      `yaml:"port"`
		Name             string   `yaml:"name"`
		Params           []string `yaml:"params"`
		ConnectionString string
	} `yaml:"database"`
}

func RetrieveConfig() {
	config, err := retrieveConfig()
	if err != nil {
		logger.Panicf("failed to get config: %s", err)
	}

	if err != nil {
		logger.Panicf("failed to get config from config server: %s", err)
	}

	config.Database.ConnectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name, strings.Join(config.Database.Params, "&"))
	MyConfig = config
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
