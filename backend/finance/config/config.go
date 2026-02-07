package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"reflect"
	"sen1or/letslive/finance/constants"
	"sen1or/letslive/finance/pkg/discovery"
	"sen1or/letslive/finance/pkg/logger"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
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

type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	Tracer   `yaml:"tracer"`
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` // in milli-second
}

type ConfigManager struct {
	ctx           context.Context
	registry      discovery.Registry
	currentConfig atomic.Pointer[Config]
	ticker        *time.Ticker
	stopChan      chan struct{}
	serviceName   string
	profile       string
}

func NewConfigManager(ctx context.Context, registry discovery.Registry, serviceName string, profile string) (*ConfigManager, error) {
	if profile == "" {
		logger.Warnf(ctx, "CONFIG_SERVER_PROFILE environment variable not set, using default 'default'")
		profile = "default"
	}

	if serviceName == "" {
		return nil, fmt.Errorf("application name cannot be empty")
	}

	cm := &ConfigManager{
		ctx:         ctx,
		registry:    registry,
		stopChan:    make(chan struct{}),
		serviceName: serviceName,
		profile:     profile,
	}

	logger.Infof(ctx, "attempting initial configuration fetch for %s-%s...", serviceName, profile)

	var initialConfig *Config
	var err error
	retryDelay := 5 * time.Second

	for {
		initialConfig, err = cm.fetchAndParseConfig()
		if err == nil {
			logger.Infof(ctx, "successfully fetched initial configuration.")
			break
		}

		logger.Errorf(ctx, "failed to fetch initial config: %v - retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)
		retryDelay *= 2
		if retryDelay > 1*time.Minute {
			retryDelay = 1 * time.Minute
		}
	}

	cm.currentConfig.Store(initialConfig)

	reloadIntervalString := os.Getenv("CONFIG_SERVER_RELOAD_INTERVAL")
	reloadInterval, err := strconv.Atoi(reloadIntervalString)
	if err != nil || reloadInterval < 0 {
		reloadInterval = constants.CONFIG_SERVER_DEFAULT_RELOAD_INTERVAL
	}

	if reloadInterval > 0 {
		cm.ticker = time.NewTicker(time.Duration(reloadInterval) * time.Millisecond)
		go cm.startReloader()
		logger.Infof(ctx, "started configuration reloader with interval %v", reloadInterval)
	} else {
		logger.Infof(ctx, "configuration reloading disabled (interval <= 0)")
	}

	return cm, nil
}

func (cm *ConfigManager) GetConfig() *Config {
	return cm.currentConfig.Load()
}

func (cm *ConfigManager) Stop() {
	if cm.ticker != nil {
		logger.Infof(cm.ctx, "stopping configuration reloader...")
		cm.ticker.Stop()
		close(cm.stopChan)
		logger.Infof(cm.ctx, "configuration reloader stopped.")
	}
}

func (cm *ConfigManager) startReloader() {
	for {
		select {
		case <-cm.ticker.C:
			logger.Debugf(cm.ctx, "polling for configuration changes...")
			cm.reload()
		case <-cm.stopChan:
			logger.Debugf(cm.ctx, "exiting configuration reload loop.")
			return
		}
	}
}

func (cm *ConfigManager) reload() {
	newConfig, err := cm.fetchAndParseConfig()
	if err != nil {
		logger.Errorf(cm.ctx, "failed to fetch/parse config during reload: %v", err)
		return
	}

	currentConfig := cm.GetConfig()
	if !reflect.DeepEqual(currentConfig, newConfig) {
		cm.currentConfig.Store(newConfig)
		logger.Infof(cm.ctx, "configuration updated.")
	} else {
		logger.Debugf(cm.ctx, "configuration unchanged.")
	}
}

func (cm *ConfigManager) fetchAndParseConfig() (*Config, error) {
	configserverURL, err := cm.registry.ServiceAddress(context.Background(), "configserver")
	if err != nil {
		return nil, fmt.Errorf("failed to discover configserver: %w", err)
	}

	url := fmt.Sprintf("%s/%s-%s.yml", configserverURL, cm.serviceName, cm.profile)
	logger.Debugf(cm.ctx, "fetching config from: %s", url)

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

	dbUser := os.Getenv("FINANCE_DB_USER")
	dbPassword := os.Getenv("FINANCE_DB_PASSWORD")
	if dbUser == "" || dbPassword == "" {
		logger.Warnf(cm.ctx, "database credentials (FINANCE_DB_USER, FINANCE_DB_PASSWORD) not found in environment.")
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

	return &config, nil
}
