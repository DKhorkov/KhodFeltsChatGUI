package common_test

import (
	"os"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/common"
)

func TestCreateLogsDir(t *testing.T) {
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
				os.RemoveAll(common.LogsDir)
			},
			cleanup: func() {
				// Удаляем созданную директорию
				os.RemoveAll(common.LogsDir)
			},
			expectedExist: true,
		},
		{
			name: "logs directory already exists - should not fail",
			setup: func() {
				// Создаем директорию заранее
				os.MkdirAll(common.LogsDir, os.ModePerm)
			},
			cleanup: func() {
				// Удаляем директорию после теста
				os.RemoveAll(common.LogsDir)
			},
			expectedExist: true,
		},
		{
			name: "logs directory exists as file - should not fail",
			setup: func() {
				// Создаем файл вместо директории
				os.RemoveAll(common.LogsDir)
				f, _ := os.Create(common.LogsDir)
				f.Close()
			},
			cleanup: func() {
				// Удаляем файл после теста
				os.RemoveAll(common.LogsDir)
			},
			expectedExist: true, // функция не удаляет файл, просто пытается создать директорию
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tests := []struct {
		name     string
		constant any
		expected any
	}{
		{
			name:     "LoggingTraceSkipLevel constant",
			constant: common.LoggingTraceSkipLevel,
			expected: 1,
		},
		{
			name:     "LogsDir constant",
			constant: common.LogsDir,
			expected: "logs",
		},
		{
			name:     "LogsPath format",
			constant: common.LogsPath,
			expected: "logs/%s.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
