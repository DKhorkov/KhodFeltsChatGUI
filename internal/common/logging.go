package common

import (
	"os"
	"path/filepath"
)

const LoggingTraceSkipLevel = 1

var (
	LogsDir  = filepath.Join(AppDataDir(), "logs")
	LogsPath = filepath.Join(LogsDir, "%s.log")
)

func CreateLogsDir() {
	if _, err := os.Stat(LogsDir); os.IsNotExist(err) {
		_ = os.MkdirAll(LogsDir, os.ModePerm)
	}
}
