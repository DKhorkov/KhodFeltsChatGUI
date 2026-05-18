package common

import (
	"os"
	"path/filepath"
)

const appName = "KhodFeltsChatGUI"

// AppDataDir возвращает абсолютный путь к директории данных приложения.
// На macOS: ~/Library/Application Support/KhodFeltsChatGUI
// На Linux: ~/.config/KhodFeltsChatGUI
// На Windows: %AppData%/KhodFeltsChatGUI.
func AppDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback на домашнюю директорию
		home, _ := os.UserHomeDir()

		return filepath.Join(home, "."+appName)
	}

	return filepath.Join(configDir, appName)
}

// CreateAppDataDir создаёт директорию данных приложения, если она не существует.
func CreateAppDataDir() {
	dir := AppDataDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, os.ModePerm)
	}
}
