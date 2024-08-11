package config

import "time"

var (
	RefreshTokenExpiresDuration = 7 * 24 * time.Hour
	AccessTokenExpiresDuration  = 5 * time.Minute

	RefreshTokenMaxAge = 7 * 24 * 3600
	AccessTokenMaxAge  = 5 * 60
)
