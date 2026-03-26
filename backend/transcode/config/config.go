package config

import (
	"fmt"
	neturl "net/url"
	"os"
	"strings"
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

type Database struct {
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Name             string   `yaml:"name"`
	Params           []string `yaml:"params"`
	ConnectionString string
}

type Config struct {
	Service   `yaml:"service"`
	RTMP      `yaml:"rtmp"`
	Transcode `yaml:"transcode"`
	MinIO     `yaml:"minio"`
	Database  `yaml:"database"`
	Webserver struct {
		Port int `yaml:"port"`
	} `yaml:"webserver"`
}

func PostProcess(config *Config) error {
	if config.Database.Host != "" {
		dbUser := os.Getenv("LIVESTREAM_DB_USER")
		dbPassword := os.Getenv("LIVESTREAM_DB_PASSWORD")

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
	}

	return nil
}
