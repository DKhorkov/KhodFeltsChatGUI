package reactions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/kfcGUI/internal/repositories/base"
)

const accessTokenCookieName = "accessToken"

type Repository struct {
	base.Repository

	httpClient interfaces.HTTPClient
	baseURL    string
	mu         sync.RWMutex
}

func New(httpClient interfaces.HTTPClient, baseURL string) *Repository {
	return &Repository{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// ListReactions — публичный роут, cookie с accessToken не отправляем.
func (r *Repository) ListReactions(ctx context.Context) ([]domains.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := r.baseURL + "/reactions"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.CloseBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(data))
	}

	var reactions []domains.Reaction
	if err = json.Unmarshal(data, &reactions); err != nil {
		return nil, err
	}

	return reactions, nil
}

func (r *Repository) AddMessageReaction(
	ctx context.Context,
	accessToken string,
	dto domains.MessageReactionDTO,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	body, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/messages/%d/reactions", r.baseURL, dto.MessageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)
	req.AddCookie(&http.Cookie{Name: accessTokenCookieName, Value: accessToken})

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.CloseBody(ctx, resp.Body)

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusConflict:
		return customerrors.ErrReactionAlreadyExists
	case http.StatusNotFound:
		return customerrors.ErrReactionNotFound
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return errors.New(string(data))
}

func (r *Repository) RemoveMessageReaction(
	ctx context.Context,
	accessToken string,
	dto domains.MessageReactionDTO,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := fmt.Sprintf(
		"%s/messages/%d/reactions/%d",
		r.baseURL, dto.MessageID, dto.ReactionID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, http.NoBody)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: accessTokenCookieName, Value: accessToken})

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.CloseBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(data))
	}

	return nil
}
