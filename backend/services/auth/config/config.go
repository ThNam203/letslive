package config

import "time"

// TODO: make in to a yaml config file
const (
	SERVICE_NAME     = "auth"
	AUTH_SERVER_HOST = "127.0.0.1"
	AUTH_SERVER_PORT = "7777"

	REGISTRY_ADDR = "192.168.0.3:8500"

	REFRESH_TOKEN_EXPIRES_DURATION = 7 * 24 * time.Hour
	ACCESS_TOKEN_EXPIRES_DURATION  = 5 * time.Minute

	REFRESH_TOKEN_MAX_AGE = 7 * 24 * 3600
	ACCESS_TOKEN_MAX_AGE  = 5 * 60

	SERVER_CRT_FILE = "server/server.crt"
	SERVER_KEY_FILE = "server/server.key"
)
