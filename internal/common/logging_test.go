package common_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/common"
)

func TestCreateLogsDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func() // подготовка перед тестом
		cleanup       func() // очистка после теста
		expectedExist bool   // ожидаемое существование директории после вызова
	}{
		{
			name: "logs directory does not exist - should create it",
			setup: func() {
				// Удаляем директорию, если она существует
				_ = os.RemoveAll(common.LogsDir)
			},
			cleanup: func() {
				// Удаляем созданную директорию
				_ = os.RemoveAll(common.LogsDir)
			},
			expectedExist: true,
		},
		{
			name: "logs directory already exists - should not fail",
			setup: func() {
				// Создаем директорию заранее
				_ = os.MkdirAll(common.LogsDir, os.ModePerm)
			},
			cleanup: func() {
				// Удаляем директорию после теста
				_ = os.RemoveAll(common.LogsDir)
			},
			expectedExist: true,
		},
		{
			name: "logs directory exists as file - should not fail",
			setup: func() {
				// Создаем файл вместо директории
				_ = os.RemoveAll(common.LogsDir)
				_ = os.MkdirAll(filepath.Dir(common.LogsDir), os.ModePerm)
				f, _ := os.Create(common.LogsDir)
				_ = f.Close()
			},
			cleanup: func() {
				// Удаляем файл после теста
				_ = os.RemoveAll(common.LogsDir)
			},
			expectedExist: true, // функция не удаляет файл, просто пытается создать директорию
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Подготовка
			if tt.setup != nil {
				tt.setup()
			}

			// Очистка после теста
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// Вызов тестируемой функции
			common.CreateLogsDir()

			// Проверка результата
			_, err := os.Stat(common.LogsDir)
			exists := !os.IsNotExist(err)

			if exists != tt.expectedExist {
				t.Errorf("CreateLogsDir() после вызова, директория существует = %v, ожидалось %v",
					exists, tt.expectedExist)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{
			name:     "LoggingTraceSkipLevel constant",
			got:      common.LoggingTraceSkipLevel,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestLogsDirIsAbsolute(t *testing.T) {
	t.Parallel()

	if !filepath.IsAbs(common.LogsDir) {
		t.Errorf("LogsDir should be absolute path, got %s", common.LogsDir)
	}
}

func TestLogsPathContainsLogsDir(t *testing.T) {
	t.Parallel()

	if !strings.HasPrefix(common.LogsPath, common.LogsDir) {
		t.Errorf("LogsPath = %s should be inside LogsDir = %s", common.LogsPath, common.LogsDir)
	}
}
