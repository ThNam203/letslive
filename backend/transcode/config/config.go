package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sen1or/letslive/transcode/constants"
	"sen1or/letslive/transcode/pkg/discovery"
	"sen1or/letslive/transcode/pkg/logger"
	"strconv"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Name            string `yaml:"name"`
	Hostname        string `yaml:"hostname"`
	APIPort         int    `yaml:"apiPort"`
	RtmpBindAddress string `yaml:"rtmpBindAddress"`
	Port            int    `yaml:"port"`
}

type RTMP struct {
	Port int `yaml:"port"`
}

type MinIO struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	BucketName string `yaml:"bucketName"`
	ReturnURL  string `yaml:"returnURL"`
}

type Transcode struct {
	PublicHLSPath        string `yaml:"publicHLSPath"`
	PrivateHLSPath       string `yaml:"privateHLSPath"`
	VODPlaybackUrlPrefix string `yaml:"vodPlaybackUrlPrefix"`

	FFMpegSetting struct {
		FFMpegPath     string `yaml:"ffmpegPath"`
		MasterFileName string `yaml:"masterFileName"`
		HLSTime        int    `yaml:"hlsTime"`
		CRF            int    `yaml:"crf"`
		Preset         string `yaml:"preset"`
		HlsListSize    int    `yaml:"hlsListSize"`
		HlsMaxSize     int    `yaml:"hlsMaxSize"`
		Qualities      []struct {
			Resolution string `yaml:"resolution"`
			MaxBitrate string `yaml:"maxBitrate"`
			FPS        int    `yaml:"fps"`
			BufSize    string `yaml:"bufSize"`
		} `yaml:"qualities"`
	} `yaml:"ffmpegSetting"`
}

type Config struct {
	Service   `yaml:"service"`
	RTMP      `yaml:"rtmp"`
	Transcode `yaml:"transcode"`
	MinIO     `yaml:"minio"`
	Webserver struct {
		Port int `yaml:"port"`
	} `yaml:"webserver"`
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
		cm.ticker = time.NewTicker(time.Duration(reloadInterval) * time.Millisecond)
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

	return &config, nil
}
