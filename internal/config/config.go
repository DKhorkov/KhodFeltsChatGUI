package config

import (
	"fmt"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
)

func New() Config {
	return Config{
		HTTP: HTTPConfig{
			Timeout: time.Second * time.Duration(
				loadenv.GetEnvAsInt("HTTP_CLIENT_TIMEOUT", 5),
			),
			WebsocketURL: loadenv.GetEnv("HTTP_WEBSOCKET_URL", "ws://185.119.59.215:8080"),
			BaseURL:      loadenv.GetEnv("HTTP_BASE_URL", "http://185.119.59.215:8080"),
		},
		Logging: logging.Config{
			Level: logging.Levels.DEBUG,
			LogFilePath: fmt.Sprintf(
				common.LogsPath,
				time.Now().In(common.Timezone).Format(common.DateFormat),
			),
		},
	}
}

type Config struct {
	HTTP    HTTPConfig
	Logging logging.Config
}

type HTTPConfig struct {
	Timeout      time.Duration
	WebsocketURL string
	BaseURL      string
}
