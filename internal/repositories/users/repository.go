package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
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

	avatarFormFileKey = "avatar"
	avatarFileName    = "avatar.jpg"
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

func (r *Repository) UpdateUser(
	ctx context.Context,
	accessToken string,
	updateUserData domains.UpdateUserDTO,
) (*domains.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	body, err := json.Marshal(updateUserData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		r.baseURL+"/users/me",
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(data))
	}

	var updatedUser domains.User
	if err = json.Unmarshal(data, &updatedUser); err != nil {
		return nil, err
	}

	return &updatedUser, nil
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
			queryParams.Add(limitQueryParamName, strconv.FormatUint(*pagination.Limit, 10))
		}

		if pagination.Offset != nil {
			queryParams.Add(offsetQueryParamName, strconv.FormatUint(*pagination.Offset, 10))
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

func (r *Repository) UpdateAvatar(
	ctx context.Context,
	accessToken string,
	fileData []byte,
) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(avatarFormFileKey, avatarFileName)
	if err != nil {
		return "", err
	}

	if _, err = part.Write(fileData); err != nil {
		return "", err
	}

	if err = writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, r.baseURL+"/users/me/avatar", &body)
	if err != nil {
		return "", err
	}

	req.Header.Set(common.ContentTypeHeaderName, writer.FormDataContentType())

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.CloseBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(data))
	}

	return string(data), nil
}

func (r *Repository) DeleteAvatar(
	ctx context.Context,
	accessToken string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.baseURL+"/users/me/avatar", http.NoBody)
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
