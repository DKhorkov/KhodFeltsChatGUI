package tokens

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
	permission = 0o600

	tokensFilename = "tokens.json"

	// JSON view variables.
	prefix = ""
	indent = "  "
)

var tokensFilePath = filepath.Join(common.AppDataDir(), tokensFilename)

type Repository struct {
	mu sync.RWMutex
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) Save(_ context.Context, tokens domains.TokensDTO) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.MarshalIndent(tokens, prefix, indent)
	if err != nil {
		return err
	}

	return os.WriteFile(tokensFilePath, data, permission)
}

func (r *Repository) Load(_ context.Context) (*domains.TokensDTO, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(tokensFilePath)
	if err != nil {
		return nil, err
	}

	var tokens domains.TokensDTO
	if err = json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

func (r *Repository) Delete(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return os.Remove(tokensFilePath)
}
