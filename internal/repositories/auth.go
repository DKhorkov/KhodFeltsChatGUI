package repositories

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
)

var (
	accessTokenCookieName  = "accessToken"
	refreshTokenCookieName = "refreshToken"
)

type AuthRepository struct {
	Repository

	httpClient interfaces.HTTPClient
	baseURL    string
	mu         sync.Mutex
}

func NewAuthRepository(
	httpClient interfaces.HTTPClient,
	baseURL string,
) *AuthRepository {
	return &AuthRepository{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *AuthRepository) Register(
	ctx context.Context,
	registerData domains.RegisterDTO,
) (*domains.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	body, err := json.Marshal(registerData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.baseURL+"/users",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.closeBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(string(data))
	}

	var createdUser domains.User
	if err = json.Unmarshal(data, &createdUser); err != nil {
		return nil, err
	}

	return &createdUser, nil
}

func (r *AuthRepository) Login(
	ctx context.Context,
	email, password string,
) (*domains.TokensDTO, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	input := domains.LoginDTO{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.baseURL+"/sessions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(data))
	}

	var tokens domains.TokensDTO

	for _, cookie := range resp.Cookies() {
		if cookie.Name == accessTokenCookieName {
			tokens.AccessToken = cookie.Value
		}

		if cookie.Name == refreshTokenCookieName {
			tokens.RefreshToken = cookie.Value
		}
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return nil, fmt.Errorf(
			"%w: %s",
			customerrors.ErrLogin,
			fmt.Sprintf("failed to get tokens from cookies. Tokens: %v", tokens),
		)
	}

	return &tokens, nil
}

func (r *AuthRepository) Logout(ctx context.Context, accessToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		r.baseURL+"/sessions",
		http.NoBody,
	)
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

	defer r.closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(data))
	}

	return nil
}

func (r *AuthRepository) RefreshTokens(
	ctx context.Context,
	refreshToken string,
) (*domains.TokensDTO, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		r.baseURL+"/sessions",
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.AddCookie(
		&http.Cookie{
			Name:  refreshTokenCookieName,
			Value: refreshToken,
		},
	)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(data))
	}

	var tokens domains.TokensDTO

	for _, cookie := range resp.Cookies() {
		if cookie.Name == accessTokenCookieName {
			tokens.AccessToken = cookie.Value
		}

		if cookie.Name == refreshTokenCookieName {
			tokens.RefreshToken = cookie.Value
		}
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return nil, fmt.Errorf(
			"%w: %s",
			customerrors.ErrRefreshTokens,
			fmt.Sprintf("failed to get tokens from cookies. Tokens: %v", tokens),
		)
	}

	return &tokens, nil
}

func (r *AuthRepository) SendVerifyEmail(ctx context.Context, email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	input := domains.SendVerifyEmailMessageDTO{
		Email: email,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.baseURL+"/users/email/verify",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer r.closeBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(data))
	}

	return nil
}
