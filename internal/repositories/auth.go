package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/libs/logging"
)

var (
	accessTokenCookieName  = "accessToken"
	refreshTokenCookieName = "refreshToken"
)

type AuthRepository struct {
	authConfig config.AuthConfig
	logger     logging.Logger
	httpClient *http.Client
}

func New(
	authConfig config.AuthConfig,
	logger logging.Logger,
	httpConfig config.HTTPConfig,
) *AuthRepository {
	return &AuthRepository{
		authConfig: authConfig,
		logger:     logger,
		httpClient: &http.Client{Timeout: httpConfig.Timeout},
	}
}

func (r *AuthRepository) Register(ctx context.Context, user domains.User) (*domains.User, error) {
	input := domains.RegisterDTO{
		Email:    user.Email,
		Username: user.Username,
		Password: user.Password,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.authConfig.RegisterURL,
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
		return nil, fmt.Errorf("%w: %s. Status: %s", errors.ErrRegister, data, resp.Status)
	}

	var createdUser *domains.User
	if err = json.Unmarshal(data, createdUser); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (r *AuthRepository) Login(
	ctx context.Context,
	email, password string,
) (*domains.TokensDTO, error) {
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
		r.authConfig.LoginURL,
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

		return nil, fmt.Errorf("%w: %s. Status: %s", errors.ErrLogin, data, resp.Status)
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
			errors.ErrLogin,
			fmt.Sprintf("failed to get tokens from cookies. Tokens: %v", tokens),
		)
	}

	return &tokens, nil
}

func (r *AuthRepository) Logout(ctx context.Context, accessToken string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		r.authConfig.LogoutURL,
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

		return fmt.Errorf("%w: %s. Status: %s", errors.ErrLogout, data, resp.Status)
	}

	return nil
}

func (r *AuthRepository) RefreshTokens(
	ctx context.Context,
	refreshToken string,
) (*domains.TokensDTO, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		r.authConfig.RefreshTokensURL,
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

		return nil, fmt.Errorf("%w: %s. Status: %s", errors.ErrRefreshTokens, data, resp.Status)
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
			errors.ErrRefreshTokens,
			fmt.Sprintf("failed to get tokens from cookies. Tokens: %v", tokens),
		)
	}

	return &tokens, nil
}

func (r *AuthRepository) closeBody(ctx context.Context, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		logging.LogErrorContext(
			ctx,
			r.logger,
			"failed to close response body",
			err,
		)
	}
}
