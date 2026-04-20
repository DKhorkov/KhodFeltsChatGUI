package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

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

func New(
	httpClient interfaces.HTTPClient,
	baseURL string,
) *Repository {
	return &Repository{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *Repository) GetCurrentUser(
	ctx context.Context,
	accessToken string,
) (*domains.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.baseURL+"/users/me", http.NoBody)
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

	var user domains.User
	if err = json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) SearchUsers(
	ctx context.Context,
	filters *domains.UsersFilters,
	pagination *domains.Pagination,
) ([]domains.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	queryParams := url.Values{}

	if filters != nil && filters.Username != nil {
		queryParams.Add("username", *filters.Username)
	}

	if pagination != nil {
		if pagination.Limit != nil {
			queryParams.Add(limitQueryParamName, fmt.Sprintf("%d", *pagination.Limit))
		}

		if pagination.Offset != nil {
			queryParams.Add(offsetQueryParamName, fmt.Sprintf("%d", *pagination.Offset))
		}
	}

	fullURL, err := url.Parse(r.baseURL + "/users")
	if err != nil {
		return nil, err
	}

	fullURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), http.NoBody)
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

	var users []domains.User
	if err = json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}
