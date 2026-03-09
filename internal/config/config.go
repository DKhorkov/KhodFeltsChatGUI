package config

import "time"

func New() Config {
	return Config{}
}

type Config struct {
	HTTP HTTPConfig
}

type HTTPConfig struct {
	Timeout time.Duration
}
