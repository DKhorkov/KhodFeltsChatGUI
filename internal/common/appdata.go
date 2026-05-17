package common

import (
	"io"
	"os"
	"path/filepath"
)

const (
	appName     = "KhodFeltsChatGUI"
	iconFile    = "icon.png"
	srcIconPath = "assets/icon.png"
)

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

// AppIconPath возвращает путь к иконке приложения в AppData.
func AppIconPath() string {
	return filepath.Join(AppDataDir(), iconFile)
}

// CopyAppIcon копирует иконку приложения из assets в AppData.
func CopyAppIcon() {
	src, err := os.Open(srcIconPath)
	if err != nil {
		return
	}

	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(AppIconPath())
	if err != nil {
		return
	}

	defer func() {
		_ = dst.Close()
	}()

	_, _ = io.Copy(dst, src)
}
