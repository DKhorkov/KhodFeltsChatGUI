package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/stretchr/testify/assert"
)

func TestTokensRepository_Save(t *testing.T) {
	tests := []struct {
		name        string
		tokens      domains.TokensDTO
		setup       func() error
		cleanup     func()
		expectedErr bool
	}{
		{
			name: "successful save with valid tokens",
			tokens: domains.TokensDTO{
				AccessToken:  "test_access_token_123",
				RefreshToken: "test_refresh_token_456",
			},
			setup: func() error {
				// Удаляем файл перед тестом, если существует
				_ = os.Remove(tokensFilename)

				return nil
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expectedErr: false,
		},
		{
			name: "overwrite existing file successfully",
			tokens: domains.TokensDTO{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
			setup: func() error {
				// Создаем существующий файл
				existingTokens := domains.TokensDTO{
					AccessToken:  "old_token",
					RefreshToken: "old_token",
				}
				data, _ := json.MarshalIndent(existingTokens, prefix, indent)

				return os.WriteFile(tokensFilename, data, permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expectedErr: false,
		},
		{
			name: "save with empty tokens",
			tokens: domains.TokensDTO{
				AccessToken:  "",
				RefreshToken: "",
			},
			setup: func() error {
				_ = os.Remove(tokensFilename)

				return nil
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expectedErr: false,
		},
		{
			name: "save with special characters in tokens",
			tokens: domains.TokensDTO{
				AccessToken:  "token!@#$%^&*()_+{}[]",
				RefreshToken: "refresh!@#$%^&*()_+{}[]",
			},
			setup: func() error {
				_ = os.Remove(tokensFilename)

				return nil
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expectedErr: false,
		},
		{
			name: "save with very long tokens",
			tokens: domains.TokensDTO{
				AccessToken:  string(make([]byte, 10000)),
				RefreshToken: string(make([]byte, 10000)),
			},
			setup: func() error {
				_ = os.Remove(tokensFilename)

				return nil
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
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

			repo := NewTokensRepository()
			ctx := context.Background()

			err := repo.Save(ctx, tt.tokens)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				info, err := os.Stat(tokensFilename)
				assert.NoError(t, err)

				// Проверяем права доступа
				if info.Mode().Perm() != permission {
					t.Errorf("File permissions = %v, expected %v", info.Mode().Perm(), permission)
				}

				// Проверяем содержимое файла
				data, err := os.ReadFile(tokensFilename)
				assert.NoError(t, err)

				var loadedTokens domains.TokensDTO

				err = json.Unmarshal(data, &loadedTokens)
				assert.NoError(t, err)

				assert.Equal(t, tt.tokens, loadedTokens)
			}
		})
	}
}

func TestTokensRepository_Load(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() error
		cleanup     func()
		expected    *domains.TokensDTO
		expectedErr bool
	}{
		{
			name: "successful load existing tokens",
			setup: func() error {
				tokens := domains.TokensDTO{
					AccessToken:  "access_token_123",
					RefreshToken: "refresh_token_456",
				}

				data, err := json.MarshalIndent(tokens, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(tokensFilename, data, permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected: &domains.TokensDTO{
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_456",
			},
			expectedErr: false,
		},
		{
			name: "load with empty tokens in file",
			setup: func() error {
				tokens := domains.TokensDTO{
					AccessToken:  "",
					RefreshToken: "",
				}

				data, err := json.MarshalIndent(tokens, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(tokensFilename, data, permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected: &domains.TokensDTO{
				AccessToken:  "",
				RefreshToken: "",
			},
			expectedErr: false,
		},
		{
			name: "load with special characters",
			setup: func() error {
				tokens := domains.TokensDTO{
					AccessToken:  "token!@#$%^&*()",
					RefreshToken: "refresh!@#$%^&*()",
				}

				data, err := json.MarshalIndent(tokens, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(tokensFilename, data, permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected: &domains.TokensDTO{
				AccessToken:  "token!@#$%^&*()",
				RefreshToken: "refresh!@#$%^&*()",
			},
			expectedErr: false,
		},
		{
			name: "load with unicode characters",
			setup: func() error {
				tokens := domains.TokensDTO{
					AccessToken:  "токен_доступа_привет",
					RefreshToken: "токен_обновления_мир",
				}

				data, err := json.MarshalIndent(tokens, prefix, indent)
				if err != nil {
					return err
				}

				return os.WriteFile(tokensFilename, data, permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected: &domains.TokensDTO{
				AccessToken:  "токен_доступа_привет",
				RefreshToken: "токен_обновления_мир",
			},
			expectedErr: false,
		},
		{
			name: "load when file does not exist",
			setup: func() error {
				_ = os.Remove(tokensFilename)

				return nil
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "load with corrupted JSON",
			setup: func() error {
				return os.WriteFile(tokensFilename, []byte("{invalid json}"), permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "load with empty file",
			setup: func() error {
				return os.WriteFile(tokensFilename, []byte(""), permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expected:    nil,
			expectedErr: true,
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

			repo := NewTokensRepository()
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

func TestTokensRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() error
		cleanup     func()
		expectedErr bool
	}{
		{
			name: "successful delete existing file",
			setup: func() error {
				tokens := domains.TokensDTO{
					AccessToken:  "test_token",
					RefreshToken: "test_refresh",
				}
				data, _ := json.MarshalIndent(tokens, prefix, indent)

				return os.WriteFile(tokensFilename, data, permission)
			},
			cleanup: func() {
				// Файл должен быть удален, но на всякий случай удаляем
				_ = os.Remove(tokensFilename)
			},
			expectedErr: false,
		},
		{
			name: "delete when file does not exist",
			setup: func() error {
				_ = os.Remove(tokensFilename)

				return nil
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
			expectedErr: true,
		},
		{
			name: "delete after multiple saves and deletes",
			setup: func() error {
				// Создаем файл
				tokens := domains.TokensDTO{
					AccessToken:  "token1",
					RefreshToken: "refresh1",
				}

				data, _ := json.MarshalIndent(tokens, prefix, indent)
				if err := os.WriteFile(tokensFilename, data, permission); err != nil {
					return err
				}
				// Удаляем его
				if err := os.Remove(tokensFilename); err != nil {
					return err
				}
				// Создаем снова
				tokens2 := domains.TokensDTO{
					AccessToken:  "token2",
					RefreshToken: "refresh2",
				}
				data2, _ := json.MarshalIndent(tokens2, prefix, indent)

				return os.WriteFile(tokensFilename, data2, permission)
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
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

			repo := NewTokensRepository()
			ctx := context.Background()

			err := repo.Delete(ctx)

			// Проверяем, что файл действительно удален
			if !tt.expectedErr {
				assert.NoError(t, err)

				_, err := os.Stat(tokensFilename)
				if !os.IsNotExist(err) {
					t.Errorf("File still exists after delete")
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTokensRepository_Concurrency(t *testing.T) {
	tests := []struct {
		name       string
		operations []struct {
			opType string // "save", "load", "delete"
			tokens domains.TokensDTO
		}
		cleanup func()
	}{
		{
			name: "concurrent save operations",
			operations: []struct {
				opType string
				tokens domains.TokensDTO
			}{
				{"save", domains.TokensDTO{AccessToken: "token1", RefreshToken: "refresh1"}},
				{"save", domains.TokensDTO{AccessToken: "token2", RefreshToken: "refresh2"}},
				{"save", domains.TokensDTO{AccessToken: "token3", RefreshToken: "refresh3"}},
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
		},
		{
			name: "concurrent save and load operations",
			operations: []struct {
				opType string
				tokens domains.TokensDTO
			}{
				{"save", domains.TokensDTO{AccessToken: "token1", RefreshToken: "refresh1"}},
				{"load", domains.TokensDTO{}},
				{"save", domains.TokensDTO{AccessToken: "token2", RefreshToken: "refresh2"}},
				{"load", domains.TokensDTO{}},
			},
			cleanup: func() {
				_ = os.Remove(tokensFilename)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Cleanup before test
			_ = os.Remove(tokensFilename)

			// Cleanup after test
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			repo := NewTokensRepository()
			ctx := context.Background()

			var wg sync.WaitGroup
			for _, op := range tt.operations {
				wg.Add(1)

				go func(opType string, tokens domains.TokensDTO) {
					defer wg.Done()

					switch opType {
					case "save":
						_ = repo.Save(ctx, tokens)
					case "load":
						_, _ = repo.Load(ctx)
					case "delete":
						_ = repo.Delete(ctx)
					}
				}(op.opType, op.tokens)
			}

			wg.Wait()
		})
	}
}
