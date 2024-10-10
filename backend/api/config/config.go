package config

import "time"

const (
	REFRESH_TOKEN_EXPIRES_DURATION = 7 * 24 * time.Hour
	ACCESS_TOKEN_EXPIRES_DURATION  = 5 * time.Minute

	REFRESH_TOKEN_MAX_AGE = 7 * 24 * 3600
	ACCESS_TOKEN_MAX_AGE  = 5 * 60

	SERVER_CRT_FILE = "server/server.crt"
	SERVER_KEY_FILE = "server/server.key"
)
