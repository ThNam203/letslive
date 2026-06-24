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

// Deposit holds amount bounds (in minor units / cents) for the POST /deposits flow.
type Deposit struct {
	MinAmount int64 `yaml:"minAmount"`
	MaxAmount int64 `yaml:"maxAmount"`
}

// Stripe holds checkout-session redirect URLs and is augmented at runtime
// with the API key and webhook secret loaded from environment variables.
type Stripe struct {
	SuccessURL       string `yaml:"successUrl"`
	CancelURL        string `yaml:"cancelUrl"`
	FiatCurrencyCode string `yaml:"fiatCurrencyCode"`
	APIKey           string
	WebhookSecret    string
}

type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	Tracer   `yaml:"tracer"`
	Deposit  `yaml:"deposit"`
	Stripe   `yaml:"stripe"`
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

	config.Stripe.APIKey = os.Getenv("FINANCE_STRIPE_API_KEY")
	config.Stripe.WebhookSecret = os.Getenv("FINANCE_STRIPE_WEBHOOK_SECRET")

	if config.Stripe.FiatCurrencyCode == "" {
		return fmt.Errorf("stripe.fiatCurrencyCode must be set in config (e.g. \"usd\")")
	}

	return nil
}
