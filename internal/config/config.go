package config

import "time"

func New() Config {
	return Config{}
}

type Config struct {
	Auth AuthConfig
	HTTP HTTPConfig
}

type AuthConfig struct {
	RegisterURL      string
	LoginURL         string
	LogoutURL        string
	RefreshTokensURL string
}

type HTTPConfig struct {
	Timeout time.Duration
}
