package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/kfcGUI/internal/repositories/base"
)

const (
	accessTokenCookieName = "accessToken"

	limitQueryParamName  = "limit"
	offsetQueryParamName = "offset"
)

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

func (r *Repository) GetChatMessages(
	ctx context.Context,
	accessToken string,
	chatID uint64,
	pagination *domains.Pagination,
) ([]domains.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	queryParams := url.Values{}

	if pagination != nil {
		if pagination.Limit != nil {
			queryParams.Add(limitQueryParamName, strconv.FormatUint(*pagination.Limit, 10))
		}

		if pagination.Offset != nil {
			queryParams.Add(offsetQueryParamName, strconv.FormatUint(*pagination.Offset, 10))
		}
	}

	path := fmt.Sprintf("%s/chats/%d/messages", r.baseURL, chatID)

	fullURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	fullURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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

	var messages []domains.Message
	if err = json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *Repository) GetMessageByID(
	ctx context.Context,
	accessToken string,
	messageID uint64,
) (*domains.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := fmt.Sprintf("%s/messages/%d", r.baseURL, messageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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

	var message domains.Message
	if err = json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *Repository) UpdateMessage(
	ctx context.Context,
	accessToken string,
	dto domains.UpdateMessageDTO,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	body, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/messages/%d", r.baseURL, dto.MessageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)
	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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

func (r *Repository) ListReactions(
	ctx context.Context,
	accessToken string,
) ([]domains.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := fmt.Sprintf("%s/reactions", r.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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
	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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

func (r *Repository) DeleteMessage(
	ctx context.Context,
	accessToken string,
	dto domains.DeleteMessageDTO,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	body, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/messages/%d", r.baseURL, dto.MessageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)
	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

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
