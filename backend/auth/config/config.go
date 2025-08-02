package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sen1or/letslive/auth/constants"
	"sen1or/letslive/auth/pkg/discovery"
	"sen1or/letslive/auth/pkg/logger"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

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

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` /// in milli-second
}

type Config struct {
	Service      `yaml:"service"`
	JWT          `yaml:"jwt"`
	Database     `yaml:"database"`
	Verification `yaml:"verification"`
	Tracer       `yaml:"tracer"`
}

type ConfigManager struct {
	registry      discovery.Registry
	currentConfig atomic.Pointer[Config] // Stores *Config
	ticker        *time.Ticker
	stopChan      chan struct{}
	serviceName   string
	profile       string
}

// NewConfigManager creates a new ConfigManager, performs the initial fetch with retry,
// and starts the background reloader.
func NewConfigManager(registry discovery.Registry, serviceName string, profile string) (*ConfigManager, error) {
	if profile == "" {
		logger.Warnf("CONFIG_SERVER_PROFILE environment variable not set, using default 'default'")
		profile = "default"
	}

	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	cm := &ConfigManager{
		registry:    registry,
		stopChan:    make(chan struct{}),
		serviceName: serviceName,
		profile:     profile,
	}

	logger.Infof("attempting initial configuration fetch for %s-%s...", serviceName, profile)

	// initial fetch with infinite retry
	var initialConfig *Config
	var err error
	retryDelay := 5 * time.Second

	for {
		initialConfig, err = cm.fetchAndParseConfig()
		if err == nil {
			logger.Infof("successfully fetched initial configuration.")
			break
		}

		logger.Errorf("failed to fetch initial config: %v - retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)
		retryDelay *= 2
		if retryDelay > 1*time.Minute {
			retryDelay = 1 * time.Minute
		}
	}

	// store the successfully fetched initial configuration
	cm.currentConfig.Store(initialConfig)

	var reloadIntervalString = os.Getenv("CONFIG_SERVER_RELOAD_INTERVAL") // milli
	reloadInterval, err := strconv.Atoi(reloadIntervalString)
	if err != nil || reloadInterval < 0 {
		reloadInterval = constants.CONFIG_SERVER_DEFAULT_RELOAD_INTERVAL
	}

	// start background polling for updates
	if reloadInterval > 0 {
		cm.ticker = time.NewTicker(time.Duration(reloadInterval * int(time.Millisecond)))
		go cm.startReloader()
		logger.Infof("started configuration reloader with interval %v", reloadInterval)
	} else {
		logger.Infof("configuration reloading disabled (interval <= 0)")
	}

	return cm, nil
}

func (cm *ConfigManager) GetConfig() *Config {
	return cm.currentConfig.Load()
}

// stop halts the background configuration reloader.
func (cm *ConfigManager) Stop() {
	if cm.ticker != nil {
		logger.Infof("stopping configuration reloader...")
		cm.ticker.Stop()
		close(cm.stopChan) // Signal the goroutine to exit
		logger.Infof("configuration reloader stopped.")
	}
}

// startReloader runs the polling loop in a separate goroutine.
func (cm *ConfigManager) startReloader() {
	for {
		select {
		case <-cm.ticker.C:
			logger.Debugf("polling for configuration changes...")
			cm.reload()
		case <-cm.stopChan:
			logger.Debugf("exiting configuration reload loop.")
			return // Exit the goroutine
		}
	}
}

// reload fetches the latest configuration and updates it if changed.
func (cm *ConfigManager) reload() {
	newConfig, err := cm.fetchAndParseConfig()
	if err != nil {
		logger.Errorf("failed to fetch/parse config during reload: %v", err)
		return
	}

	currentConfig := cm.GetConfig()

	// compare the new config with the current one
	if !reflect.DeepEqual(currentConfig, newConfig) {
		cm.currentConfig.Store(newConfig)
		logger.Infof("configuration updated.")
		// trigger actions here ##consider channels or callbacks.
	} else {
		logger.Debugf("configuration unchanged.")
	}
}

// fetchAndParseConfig handles the logic of getting the config from the server and parsing it.
// This replaces the old retrieveServiceConfig function.
func (cm *ConfigManager) fetchAndParseConfig() (*Config, error) {
	// get config server from service discovery
	configserverURL, err := cm.registry.ServiceAddress(context.Background(), "configserver")
	if err != nil {
		return nil, fmt.Errorf("failed to discover configserver: %w", err)
	}

	// construct config-server url (Spring Cloud Config Server format for label/profile/app)
	// url format: /{application}-{profile}.yml  (or .json, .properties)
	// or /{label}/{application}-{profile}.yml if using labels (e.g., git branches)
	url := fmt.Sprintf(
		"%s/%s-%s.yml",
		configserverURL,
		cm.serviceName,
		cm.profile,
	)
	logger.Debugf("fetching config from: %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to config server (%s): %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("config server returned non-OK status: %d - %s, body: %s", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from config server: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("received empty response body from config server")
	}

	var config Config
	err = yaml.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config YAML: %w", err)
	}

	dbUser := os.Getenv("AUTH_DB_USER")
	dbPassword := os.Getenv("AUTH_DB_PASSWORD")
	if dbUser == "" || dbPassword == "" {
		logger.Warnf("database credentials (AUTH_DB_USER, AUTH_DB_PASSWORD) not found in environment.")
	}

	config.Database.ConnectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s",
		dbUser,
		dbPassword,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
		strings.Join(config.Database.Params, "&"),
	)

	return &config, nil
}
