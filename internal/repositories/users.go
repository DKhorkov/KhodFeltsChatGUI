package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/errors"
)

type UsersRepository struct {
	Repository
	httpClient *http.Client
	baseURL    string
}

func NewUsersRepository(
	httpClient *http.Client,
	baseURL string,
) *UsersRepository {
	return &UsersRepository{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *UsersRepository) GetCurrentUser(ctx context.Context, accessToken string) (*domains.User, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/users/me", r.baseURL),
		nil,
	)
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
		return nil, fmt.Errorf("%w: %s. Status: %s", errors.ErrUserNotFound, data, resp.Status)
	}

	var user domains.User
	if err = json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UsersRepository) SearchUsers(ctx context.Context, username string, limit, offset int) ([]domains.User, error) {
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
		nil,
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
		return nil, fmt.Errorf("%w: %s. Status: %s", errors.ErrUserNotFound, data, resp.Status)
	}

	var users []domains.User
	if err = json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}
