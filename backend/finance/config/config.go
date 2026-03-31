package config

import (
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
	MigrationPath    string   `yaml:"migration-path"`
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Name             string   `yaml:"name"`
	Params           []string `yaml:"params"`
	ConnectionString string
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` // in milli-second
}

type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	Tracer   `yaml:"tracer"`
}

// TracerConfig interface implementation
func (c Config) GetServiceName() string      { return c.Service.Name }
func (c Config) GetTracerEndpoint() string   { return c.Tracer.Endpoint }
func (c Config) GetTracerBatchTimeout() int  { return c.Tracer.BatchTimeout }
func (c Config) IsSecure() bool              { return c.Tracer.Secure }

// PostProcess builds the database connection string from environment variables.
func PostProcess(config *Config) error {
	dbUser := os.Getenv("FINANCE_DB_USER")
	dbPassword := os.Getenv("FINANCE_DB_PASSWORD")

	dbURL := &neturl.URL{
		Scheme: "postgres",
		User:   neturl.UserPassword(dbUser, dbPassword),
		Host:   fmt.Sprintf("%s:%d", config.Database.Host, config.Database.Port),
		Path:   "/" + config.Database.Name,
	}
	if len(config.Database.Params) > 0 {
		dbURL.RawQuery = strings.Join(config.Database.Params, "&")
	}
	config.Database.ConnectionString = dbURL.String()

	return nil
}
