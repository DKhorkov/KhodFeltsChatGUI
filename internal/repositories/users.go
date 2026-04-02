package repositories

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
)

type UsersRepository struct {
	Repository

	httpClient interfaces.HTTPClient
	baseURL    string
	mu         sync.RWMutex
}

func NewUsersRepository(
	httpClient interfaces.HTTPClient,
	baseURL string,
) *UsersRepository {
	return &UsersRepository{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *UsersRepository) GetCurrentUser(
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
	defer r.closeBody(ctx, resp.Body)

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

func (r *UsersRepository) SearchUsers(
	ctx context.Context,
	username string,
	limit, offset int,
) ([]domains.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			"%s/users?username=%s&limit=%d&offset=%d",
			r.baseURL,
			url.QueryEscape(username),
			limit,
			offset,
		),
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.closeBody(ctx, resp.Body)

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
