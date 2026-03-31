package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"strconv"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultReloadInterval = 60000 // milli

// PostProcessFunc is called after YAML unmarshaling to allow service-specific
// processing (e.g., building database connection strings from env vars).
type PostProcessFunc[T any] func(config *T) error

// ConfigManager is a generic configuration manager that fetches config from
// a config server via service discovery and supports hot-reloading.
type ConfigManager[T any] struct {
	ctx           context.Context
	registry      discovery.Registry
	currentConfig atomic.Pointer[T]
	ticker        *time.Ticker
	stopChan      chan struct{}
	serviceName   string
	profile       string
	postProcess   PostProcessFunc[T]
}

// NewConfigManager creates a new ConfigManager, performs the initial fetch with retry,
// and starts the background reloader.
func NewConfigManager[T any](ctx context.Context, registry discovery.Registry, serviceName string, profile string, postProcess PostProcessFunc[T]) (*ConfigManager[T], error) {
	if profile == "" {
		logger.Warnf(ctx, "CONFIG_SERVER_PROFILE environment variable not set, using default 'default'")
		profile = "default"
	}

	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	cm := &ConfigManager[T]{
		ctx:         ctx,
		registry:    registry,
		stopChan:    make(chan struct{}),
		serviceName: serviceName,
		profile:     profile,
		postProcess: postProcess,
	}

	logger.Infof(ctx, "attempting initial configuration fetch for %s-%s...", serviceName, profile)

	// initial fetch with infinite retry
	var initialConfig *T
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

	// store the successfully fetched initial configuration
	cm.currentConfig.Store(initialConfig)

	var reloadIntervalString = os.Getenv("CONFIG_SERVER_RELOAD_INTERVAL") // milli
	reloadInterval, err := strconv.Atoi(reloadIntervalString)
	if err != nil || reloadInterval < 0 {
		reloadInterval = DefaultReloadInterval
	}

	// start background polling for updates
	if reloadInterval > 0 {
		cm.ticker = time.NewTicker(time.Duration(reloadInterval) * time.Millisecond)
		go cm.startReloader()
		logger.Infof(ctx, "started configuration reloader with interval %v", reloadInterval)
	} else {
		logger.Infof(ctx, "configuration reloading disabled (interval <= 0)")
	}

	return cm, nil
}

func (cm *ConfigManager[T]) GetConfig() *T {
	return cm.currentConfig.Load()
}

// Stop halts the background configuration reloader.
func (cm *ConfigManager[T]) Stop() {
	if cm.ticker != nil {
		logger.Infof(cm.ctx, "stopping configuration reloader...")
		cm.ticker.Stop()
		close(cm.stopChan)
		logger.Infof(cm.ctx, "configuration reloader stopped.")
	}
}

// startReloader runs the polling loop in a separate goroutine.
func (cm *ConfigManager[T]) startReloader() {
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

// reload fetches the latest configuration and updates it if changed.
func (cm *ConfigManager[T]) reload() {
	newConfig, err := cm.fetchAndParseConfig()
	if err != nil {
		logger.Errorf(cm.ctx, "failed to fetch/parse config during reload: %v", err)
		return
	}

	currentConfig := cm.GetConfig()

	// compare the new config with the current one
	if !reflect.DeepEqual(currentConfig, newConfig) {
		cm.currentConfig.Store(newConfig)
		logger.Infof(cm.ctx, "configuration updated.")
	} else {
		logger.Debugf(cm.ctx, "configuration unchanged.")
	}
}

// fetchAndParseConfig handles the logic of getting the config from the server and parsing it.
func (cm *ConfigManager[T]) fetchAndParseConfig() (*T, error) {
	// get config server from service discovery
	configserverURL, err := cm.registry.ServiceAddress(context.Background(), "configserver")
	if err != nil {
		return nil, fmt.Errorf("failed to discover configserver: %w", err)
	}

	url := fmt.Sprintf(
		"%s/%s-%s.yml",
		configserverURL,
		cm.serviceName,
		cm.profile,
	)
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
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("config server returned non-OK status: %d - %s, and failed to read body: %w", resp.StatusCode, resp.Status, readErr)
		}
		return nil, fmt.Errorf("config server returned non-OK status: %d - %s, body: %s", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from config server: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("received empty response body from config server")
	}

	var config T
	err = yaml.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config YAML: %w", err)
	}

	if cm.postProcess != nil {
		if err := cm.postProcess(&config); err != nil {
			return nil, fmt.Errorf("error in post-processing config: %w", err)
		}
	}

	return &config, nil
}
