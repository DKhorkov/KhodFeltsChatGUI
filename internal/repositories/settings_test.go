package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/stretchr/testify/assert"
)

func TestNewSettingsRepository(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new settings repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewSettingsRepository()
			assert.NotNil(t, repo)
		})
	}
}

func TestSettingsRepository_Save(t *testing.T) {
	tests := []struct {
		name        string
		settings    domains.Settings
		setup       func() error
		cleanup     func()
		expectedErr bool
	}{
		{
			name: "successful save with light theme",
			settings: domains.Settings{
				Theme: domains.ThemeLight,
			},
			setup: func() error {
				os.Remove(settingsFilename)

				return nil
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
		{
			name: "successful save with dark theme",
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setup: func() error {
				os.Remove(settingsFilename)

				return nil
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
		{
			name: "overwrite existing file successfully",
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setup: func() error {
				// Создаем существующий файл
				existingSettings := domains.Settings{
					Theme: domains.ThemeLight,
				}
				data, _ := json.MarshalIndent(existingSettings, prefix, indent)

				return os.WriteFile(settingsFilename, data, permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
		{
			name: "save with zero value theme",
			settings: domains.Settings{
				Theme: 0, // ThemeLight по умолчанию
			},
			setup: func() error {
				os.Remove(settingsFilename)

				return nil
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Cleanup
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			repo := NewSettingsRepository()
			ctx := context.Background()

			err := repo.Save(ctx, tt.settings)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSettingsRepository_Load(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() error
		cleanup     func()
		expected    *domains.Settings
		expectedErr bool
	}{
		{
			name: "successful load with light theme",
			setup: func() error {
				settings := domains.Settings{
					Theme: domains.ThemeLight,
				}

				data, err := json.MarshalIndent(settings, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(settingsFilename, data, permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected: &domains.Settings{
				Theme: domains.ThemeLight,
			},
			expectedErr: false,
		},
		{
			name: "successful load with dark theme",
			setup: func() error {
				settings := domains.Settings{
					Theme: domains.ThemeDark,
				}

				data, err := json.MarshalIndent(settings, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(settingsFilename, data, permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected: &domains.Settings{
				Theme: domains.ThemeDark,
			},
			expectedErr: false,
		},
		{
			name: "load with default theme (0)",
			setup: func() error {
				settings := domains.Settings{
					Theme: 0,
				}

				data, err := json.MarshalIndent(settings, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(settingsFilename, data, permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected: &domains.Settings{
				Theme: domains.ThemeLight,
			},
			expectedErr: false,
		},
		{
			name: "load when file does not exist",
			setup: func() error {
				os.Remove(settingsFilename)

				return nil
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "load with corrupted JSON",
			setup: func() error {
				return os.WriteFile(settingsFilename, []byte("{invalid json}"), permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "load with empty file",
			setup: func() error {
				return os.WriteFile(settingsFilename, []byte(""), permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "load with invalid theme value",
			setup: func() error {
				// Пишем некорректное значение темы
				return os.WriteFile(settingsFilename, []byte(`{"theme": 99}`), permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expected:    &domains.Settings{Theme: 99},
			expectedErr: false, // JSON распарсится, но значение будет 99
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Cleanup
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			repo := NewSettingsRepository()
			ctx := context.Background()

			result, err := repo.Load(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSettingsRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() error
		cleanup     func()
		expectedErr bool
	}{
		{
			name: "successful delete existing file",
			setup: func() error {
				settings := domains.Settings{
					Theme: domains.ThemeDark,
				}

				data, err := json.MarshalIndent(settings, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(settingsFilename, data, permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
		{
			name: "delete when file does not exist",
			setup: func() error {
				os.Remove(settingsFilename)

				return nil
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: true,
		},
		{
			name: "delete after multiple saves and deletes",
			setup: func() error {
				// Создаем файл
				settings := domains.Settings{
					Theme: domains.ThemeLight,
				}

				data, _ := json.MarshalIndent(settings, prefix, indent)
				if err := os.WriteFile(settingsFilename, data, permission); err != nil {
					return err
				}
				// Удаляем его
				if err := os.Remove(settingsFilename); err != nil {
					return err
				}
				// Создаем снова
				settings2 := domains.Settings{
					Theme: domains.ThemeDark,
				}
				data2, _ := json.MarshalIndent(settings2, prefix, indent)

				return os.WriteFile(settingsFilename, data2, permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
		{
			name: "delete empty file",
			setup: func() error {
				return os.WriteFile(settingsFilename, []byte(""), permission)
			},
			cleanup: func() {
				os.Remove(settingsFilename)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Cleanup
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			repo := NewSettingsRepository()
			ctx := context.Background()

			err := repo.Delete(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Проверяем, что файл действительно удален
			if !tt.expectedErr {
				_, err := os.Stat(settingsFilename)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestSettingsRepository_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "settings filename constant",
			constant: settingsFilename,
			expected: "settings.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// Интеграционный тест полного цикла работы с настройками.
func TestSettingsRepository_Integration(t *testing.T) {
	t.Run("full settings lifecycle", func(t *testing.T) {
		repo := NewSettingsRepository()
		ctx := context.Background()

		// Очищаем перед тестом
		_ = os.Remove(settingsFilename)

		// 1. Сохраняем настройки (светлая тема)
		settings := domains.Settings{
			Theme: domains.ThemeLight,
		}

		err := repo.Save(ctx, settings)
		assert.NoError(t, err)

		// 2. Загружаем настройки
		loaded, err := repo.Load(ctx)
		assert.NoError(t, err)
		assert.Equal(t, settings.Theme, loaded.Theme)

		// 3. Обновляем настройки (темная тема)
		updatedSettings := domains.Settings{
			Theme: domains.ThemeDark,
		}

		err = repo.Save(ctx, updatedSettings)
		assert.NoError(t, err)

		// 4. Загружаем обновленные настройки
		loadedUpdated, err := repo.Load(ctx)
		assert.NoError(t, err)
		assert.Equal(t, updatedSettings.Theme, loadedUpdated.Theme)

		// 5. Удаляем настройки
		err = repo.Delete(ctx)
		assert.NoError(t, err)

		// 6. Проверяем, что файл удален
		_, err = repo.Load(ctx)
		assert.Error(t, err)

		// Очищаем после теста
		_ = os.Remove(settingsFilename)
	})
}
