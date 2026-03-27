package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

const (
	permission = 0o600

	tokensFilename = "tokens.json"

	// JSON view variables.
	prefix = ""
	indent = "  "
)

type TokensRepository struct {
	mu sync.RWMutex
}

func NewTokensRepository() *TokensRepository {
	return &TokensRepository{}
}

func (r *TokensRepository) Save(_ context.Context, tokens domains.TokensDTO) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.MarshalIndent(tokens, prefix, indent)
	if err != nil {
		return err
	}

	return os.WriteFile(tokensFilename, data, permission)
}

func (r *TokensRepository) Load(_ context.Context) (*domains.TokensDTO, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(tokensFilename)
	if err != nil {
		return nil, err
	}

	var tokens domains.TokensDTO
	if err = json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

func (r *TokensRepository) Delete(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return os.Remove(tokensFilename)
}
