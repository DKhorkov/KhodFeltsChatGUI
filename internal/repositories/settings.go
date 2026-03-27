package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

const (
	settingsFilename = "settings.json"
)

type SettingsRepository struct {
	mu sync.RWMutex
}

func NewSettingsRepository() *SettingsRepository {
	return &SettingsRepository{}
}

func (r *SettingsRepository) Save(_ context.Context, settings domains.Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.MarshalIndent(settings, prefix, indent)
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFilename, data, permission)
}

func (r *SettingsRepository) Load(_ context.Context) (*domains.Settings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(settingsFilename)
	if err != nil {
		return nil, err
	}

	var settings domains.Settings
	if err = json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

func (r *SettingsRepository) Delete(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return os.Remove(settingsFilename)
}
