package common_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/common"
)

func TestAppDataDirIsAbsolute(t *testing.T) {
	t.Parallel()

	dir := common.AppDataDir()
	if !filepath.IsAbs(dir) {
		t.Errorf("AppDataDir() должен возвращать абсолютный путь, получено %s", dir)
	}
}

func TestAppDataDirContainsAppName(t *testing.T) {
	t.Parallel()

	dir := common.AppDataDir()
	if filepath.Base(dir) != "KhodFeltsChatGUI" {
		t.Errorf("AppDataDir() должен заканчиваться на KhodFeltsChatGUI, получено %s", dir)
	}
}

func TestCreateAppDataDir(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(string)
		cleanup       func(string)
		expectedExist bool
	}{
		{
			name: "directory does not exist - should create it",
			setup: func(dir string) {
				_ = os.RemoveAll(dir)
			},
			cleanup: func(dir string) {
				_ = os.RemoveAll(dir)
			},
			expectedExist: true,
		},
		{
			name: "directory already exists - should not fail",
			setup: func(dir string) {
				_ = os.MkdirAll(dir, os.ModePerm)
			},
			cleanup: func(dir string) {
				_ = os.RemoveAll(dir)
			},
			expectedExist: true,
		},
		{
			name: "directory exists as file - should not fail",
			setup: func(dir string) {
				_ = os.RemoveAll(dir)
				_ = os.MkdirAll(filepath.Dir(dir), os.ModePerm)

				f, _ := os.Create(dir)
				_ = f.Close()
			},
			cleanup: func(dir string) {
				_ = os.RemoveAll(dir)
			},
			expectedExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := common.AppDataDir()

			if tt.setup != nil {
				tt.setup(dir)
			}

			if tt.cleanup != nil {
				defer tt.cleanup(dir)
			}

			common.CreateAppDataDir()

			_, err := os.Stat(dir)
			exists := !os.IsNotExist(err)

			if exists != tt.expectedExist {
				t.Errorf("CreateAppDataDir() после вызова, директория существует = %v, ожидалось %v",
					exists, tt.expectedExist)
			}
		})
	}
}
