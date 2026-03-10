package repositories

import (
	"context"
	"encoding/json"
	"os"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

const (
	permission = 0o600
	filename   = "tokens.json"

	// JSON view variables.
	prefix = ""
	indent = "  "
)

type TokensRepository struct{}

func NewTokensRepository() TokensRepository {
	return TokensRepository{}
}

func (r *TokensRepository) Save(_ context.Context, tokens *domains.TokensDTO) error {
	data, err := json.MarshalIndent(tokens, prefix, indent)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, permission)
}

func (r *TokensRepository) Load(_ context.Context) (*domains.TokensDTO, error) {
	data, err := os.ReadFile(filename)
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
	return os.Remove(filename)
}
