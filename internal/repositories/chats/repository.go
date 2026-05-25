package chats

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func (r *Repository) GetUserChats(
	ctx context.Context,
	accessToken string,
	pagination *domains.Pagination,
) ([]domains.Chat, error) {
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

	fullURL, err := url.Parse(r.baseURL + "/chats")
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

	var chats []domains.Chat
	if err = json.Unmarshal(data, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *Repository) CreateChat(
	ctx context.Context,
	accessToken string,
	chat domains.Chat,
) (*domains.Chat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	body, err := json.Marshal(chat)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.baseURL+"/chats",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	defer r.CloseBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(string(data))
	}

	var createdChat domains.Chat
	if err = json.Unmarshal(data, &createdChat); err != nil {
		return nil, err
	}

	return &createdChat, nil
}
