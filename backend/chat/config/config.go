package config

import (
	"errors"
	"fmt"
	neturl "net/url"
	"os"
	"strings"
)

type Service struct {
	Name           string `yaml:"name"`
	Hostname       string `yaml:"hostname"`
	APIBindAddress string `yaml:"apiBindAddress"`
	APIPort        int    `yaml:"apiPort"`
}

type Database struct {
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Name             string   `yaml:"name"`
	Params           []string `yaml:"params"`
	ConnectionString string   `yaml:"-"`
}

type NATS struct {
	URL string `yaml:"url"`
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` // milliseconds
}

type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	NATS     `yaml:"nats"`
	Tracer   `yaml:"tracer"`
}

func (c Config) GetServiceName() string     { return c.Service.Name }
func (c Config) GetTracerEndpoint() string  { return c.Tracer.Endpoint }
func (c Config) GetTracerBatchTimeout() int { return c.Tracer.BatchTimeout }
func (c Config) IsSecure() bool             { return c.Tracer.Secure }

// PostProcess builds the Mongo connection string from the credentials in the
// environment, the same ones the Node chat service uses.
func PostProcess(config *Config) error {
	if config.Database.Name == "" {
		return errors.New("database.name is not configured")
	}
	if config.NATS.URL == "" {
		return errors.New("nats.url is not configured")
	}

	dbURL := &neturl.URL{
		Scheme: "mongodb",
		User:   neturl.UserPassword(os.Getenv("CHAT_DB_USER"), os.Getenv("CHAT_DB_PASSWORD")),
		Host:   fmt.Sprintf("%s:%d", config.Database.Host, config.Database.Port),
		Path:   "/" + config.Database.Name,
	}
	if len(config.Database.Params) > 0 {
		dbURL.RawQuery = strings.Join(config.Database.Params, "&")
	}
	config.Database.ConnectionString = dbURL.String()

	return nil
}
