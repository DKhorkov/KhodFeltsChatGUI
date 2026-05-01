package settings

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
)

const (
	permission       = 0o600
	settingsFilename = "settings.json"

	// JSON view variables.
	prefix = ""
	indent = "  "
)

var settingsFilePath = filepath.Join(common.AppDataDir(), settingsFilename)

type Repository struct {
	mu sync.RWMutex
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) Save(_ context.Context, settings domains.Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.MarshalIndent(settings, prefix, indent)
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFilePath, data, permission)
}

func (r *Repository) Load(_ context.Context) (*domains.Settings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(settingsFilePath)
	if err != nil {
		return nil, err
	}

	var settings domains.Settings
	if err = json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

func (r *Repository) Delete(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return os.Remove(settingsFilePath)
}
