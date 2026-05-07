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
			WebsocketURL: loadenv.GetEnv("HTTP_WEBSOCKET_URL", "ws://185.119.59.215:8080/api"),
			BaseURL:      loadenv.GetEnv("HTTP_BASE_URL", "http://185.119.59.215:8080/api"),
		},
		Logging: logging.Config{
			Level: logging.Levels.DEBUG,
			LogFilePath: fmt.Sprintf(
				common.LogsPath,
				time.Now().In(common.Timezone).Format(common.DateFormat),
			),
		},
		Validation: ValidationConfig{
			EmailRegExp: loadenv.GetEnv(
				"EMAIL_REGEXP",
				"^[a-z0-9._%+\\-]+@[a-z0-9.\\-]+\\.[a-z]{2,4}$",
			),
			PasswordRegExps: loadenv.GetEnvAsSlice(
				"PASSWORD_REGEXPS",
				[]string{
					".{8,}",     // больше или равно 8 символам в длину
					"[a-z]",     // должно содержать букву на латиницу в нижнем регистре
					"[A-Z]",     // должно содержать букву на латиницу в верхнем регистре
					"[0-9]",     // должно содержать цифру
					"[^\\d\\w]", // должно содержать спецсимвол
				},
				";",
			),
			UsernameRegExps: loadenv.GetEnvAsSlice(
				"USERNAME_REGEXPS",
				[]string{
					`^.{5,70}$`,      // длина 5-70 символов
					`^[A-Za-z0-9]+$`, // только латинница и цифры
				},
				";",
			),
		},
	}
}

type Config struct {
	HTTP       HTTPConfig
	Logging    logging.Config
	Validation ValidationConfig
}

type ValidationConfig struct {
	EmailRegExp     string
	PasswordRegExps []string // Slice of rules to pass, because Go's regex doesn't support backtracking.
	UsernameRegExps []string
}

type HTTPConfig struct {
	Timeout      time.Duration
	WebsocketURL string
	BaseURL      string
}
