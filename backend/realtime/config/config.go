package config

import (
	"errors"
	"os"
)

type Service struct {
	Name           string `yaml:"name"`
	Hostname       string `yaml:"hostname"`
	APIBindAddress string `yaml:"apiBindAddress"`
	APIPort        int    `yaml:"apiPort"`
}

type NATS struct {
	URL string `yaml:"url"`
}

type WebSocket struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` // milliseconds
}

type Config struct {
	Service   `yaml:"service"`
	NATS      `yaml:"nats"`
	WebSocket `yaml:"websocket"`
	Tracer    `yaml:"tracer"`

	JWKSURL string `yaml:"-"`
}

func (c Config) GetServiceName() string     { return c.Service.Name }
func (c Config) GetTracerEndpoint() string  { return c.Tracer.Endpoint }
func (c Config) GetTracerBatchTimeout() int { return c.Tracer.BatchTimeout }
func (c Config) IsSecure() bool             { return c.Tracer.Secure }

func PostProcess(config *Config) error {
	jwksURL := os.Getenv("JWKS_URL")
	if jwksURL == "" {
		return errors.New("JWKS_URL is not set")
	}
	config.JWKSURL = jwksURL

	if config.NATS.URL == "" {
		return errors.New("nats.url is not configured")
	}
	if len(config.WebSocket.AllowedOrigins) == 0 {
		return errors.New("websocket.allowedOrigins is not configured")
	}
	return nil
}
