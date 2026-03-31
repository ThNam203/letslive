package config

import (
	"fmt"
	neturl "net/url"
	"os"
	"sen1or/letslive/shared/pkg/logger"
	"strings"
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
	Tracer   `yaml:"tracer"`
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` /// in milli-second
}

// TracerConfig interface methods
func (c Config) GetServiceName() string      { return c.Service.Name }
func (c Config) GetTracerEndpoint() string   { return c.Tracer.Endpoint }
func (c Config) GetTracerBatchTimeout() int  { return c.Tracer.BatchTimeout }
func (c Config) IsSecure() bool              { return c.Tracer.Secure }

// PostProcess builds the database connection string from environment variables.
func PostProcess(config *Config) error {
	dbUser := os.Getenv("USER_DB_USER")
	dbPassword := os.Getenv("USER_DB_PASSWORD")
	if dbUser == "" || dbPassword == "" {
		logger.Warnf(nil, "database credentials (USER_DB_USER, USER_DB_PASSWORD) not found in environment.")
	}

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
