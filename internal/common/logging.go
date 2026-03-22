package common

import "os"

const (
	LoggingTraceSkipLevel = 1
	LogsDir               = "logs"
	LogsPath              = LogsDir + "/%s.log"
)

func CreateLogsDir() {
	if _, err := os.Stat(LogsDir); os.IsNotExist(err) {
		_ = os.Mkdir(LogsDir, os.ModePerm)
	}
}
