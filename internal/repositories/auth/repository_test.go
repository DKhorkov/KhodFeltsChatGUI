package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	internalerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	mockhttp "github.com/DKhorkov/kfcGUI/mocks/http"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepository_Register(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		registerData  domains.RegisterDTO
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful registration",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				createdUser := domains.User{
					ID:             1,
					Username:       "john_doe",
					Email:          "john@example.com",
					EmailConfirmed: false,
					Password:       "",
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				userData, _ := json.Marshal(createdUser)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/users", req.URL.Path)
						assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

						var registerData domains.RegisterDTO

						body, _ := io.ReadAll(req.Body)
						_ = json.Unmarshal(body, &registerData)
						assert.Equal(t, "john_doe", registerData.Username)
						assert.Equal(t, "john@example.com", registerData.Email)
						assert.Equal(t, "password123", registerData.Password)

						return &http.Response{
							StatusCode: http.StatusCreated,
							Body:       io.NopCloser(bytes.NewReader(userData)),
						}, nil
					}).
					Times(1)
			},
			expectedUser: &domains.User{
				ID:             1,
				Username:       "john_doe",
				Email:          "john@example.com",
				EmailConfirmed: false,
				Password:       "",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			expectedError: nil,
		},
		{
			name: "registration with existing email",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusConflict,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`email already exists`)),
						),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New(`email already exists`),
		},
		{
			name: "validation error - invalid email",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`invalid email format`)),
						),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New(`invalid email format`),
		},
		{
			name: "http client error",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("connection refused"),
		},
		{
			name: "invalid json response",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusCreated,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{invalid json}`))),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: &json.SyntaxError{},
		},
		{
			name: "empty response body",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusCreated,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")

			user, err := repo.Register(context.Background(), tt.registerData)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			reflect.DeepEqual(tt.expectedUser, user)
		})
	}
}

func TestRepository_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		email          string
		password       string
		setupMocks     func(*mockhttp.MockHTTPClient)
		expectedTokens *domains.TokensDTO
		expectedError  error
	}{
		{
			name:     "successful login",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/sessions", req.URL.Path)
						assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

						var loginData domains.LoginDTO

						body, _ := io.ReadAll(req.Body)
						_ = json.Unmarshal(body, &loginData)
						assert.Equal(t, "john@example.com", loginData.Email)
						assert.Equal(t, "password123", loginData.Password)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Header: http.Header{
								"Set-Cookie": []string{
									"accessToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9; Path=/; HttpOnly",
									"refreshToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh; Path=/; HttpOnly",
								},
							},
							Body: io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedTokens: &domains.TokensDTO{
				AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh",
			},
			expectedError: nil,
		},
		{
			name:     "invalid credentials",
			email:    "wrong@example.com",
			password: "wrongpassword",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "invalid credentials"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrLogin,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "user not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrLogin,
		},
		{
			name:     "missing access token cookie",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNoContent,
						Header: http.Header{
							"Set-Cookie": []string{
								"refreshToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh; Path=/; HttpOnly",
							},
						},
						Body: io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrLogin,
		},
		{
			name:     "missing refresh token cookie",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNoContent,
						Header: http.Header{
							"Set-Cookie": []string{
								"accessToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9; Path=/; HttpOnly",
							},
						},
						Body: io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrLogin,
		},
		{
			name:     "http client error",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("network timeout")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("network timeout"),
		},
		{
			name:     "empty response body with error status",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrLogin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")

			tokens, err := repo.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedTokens, tokens)
		})
	}
}

func TestRepository_Logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:        "successful logout",
			accessToken: "valid_access_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodDelete, req.Method)
						assert.Equal(t, "/sessions", req.URL.Path)

						cookie, err := req.Cookie(accessTokenCookieName)
						assert.NoError(t, err)
						assert.Equal(t, "valid_access_token", cookie.Value)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:        "logout with invalid token",
			accessToken: "invalid_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`invalid token`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`invalid token`),
		},
		{
			name:        "logout without session",
			accessToken: "expired_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`session not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`session not found`),
		},
		{
			name:        "http client error",
			accessToken: "token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
		},
		{
			name:        "empty response body with error",
			accessToken: "token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(``),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")

			err := repo.Logout(context.Background(), tt.accessToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_RefreshTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		refreshToken   string
		setupMocks     func(*mockhttp.MockHTTPClient)
		expectedTokens *domains.TokensDTO
		expectedError  error
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPut, req.Method)
						assert.Equal(t, "/sessions", req.URL.Path)

						cookie, err := req.Cookie(refreshTokenCookieName)
						assert.NoError(t, err)
						assert.Equal(t, "valid_refresh_token", cookie.Value)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Header: http.Header{
								"Set-Cookie": []string{
									"accessToken=new_access_token_123; Path=/; HttpOnly",
									"refreshToken=new_refresh_token_456; Path=/; HttpOnly",
								},
							},
							Body: io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedTokens: &domains.TokensDTO{
				AccessToken:  "new_access_token_123",
				RefreshToken: "new_refresh_token_456",
			},
			expectedError: nil,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "invalid refresh token"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrRefreshTokens,
		},
		{
			name:         "expired refresh token",
			refreshToken: "expired_refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "token expired"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrRefreshTokens,
		},
		{
			name:         "missing access token in response",
			refreshToken: "valid_refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNoContent,
						Header: http.Header{
							"Set-Cookie": []string{
								"refreshToken=new_refresh_token_456; Path=/; HttpOnly",
							},
						},
						Body: io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrRefreshTokens,
		},
		{
			name:         "missing refresh token in response",
			refreshToken: "valid_refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNoContent,
						Header: http.Header{
							"Set-Cookie": []string{
								"accessToken=new_access_token_123; Path=/; HttpOnly",
							},
						},
						Body: io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrRefreshTokens,
		},
		{
			name:         "http client error",
			refreshToken: "refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("network error")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("network error"),
		},
		{
			name:         "server error",
			refreshToken: "refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "server error"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrRefreshTokens,
		},
		{
			name:         "empty response body with success status but no cookies",
			refreshToken: "refresh_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNoContent,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  internalerrors.ErrRefreshTokens,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")

			tokens, err := repo.RefreshTokens(context.Background(), tt.refreshToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedTokens, tokens)
		})
	}
}

func TestRepository_SendVerifyEmailMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:  "successful send verify email",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Проверяем метод
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/users/email/verify", req.URL.Path)
						assert.Equal(
							t,
							common.ApplicationJSONContentType,
							req.Header.Get(common.ContentTypeHeaderName),
						)

						// Проверяем тело запроса
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "user@example.com", input.Email)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "successful send verify email with special characters",
			email: "user.name+tag@example.co.uk",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "user.name+tag@example.co.uk", input.Email)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "send verify email to user with cyrillic domain",
			email: "user@почта.рф",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "user@почта.рф", input.Email)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "user not found - returns 404",
			email: "nonexistent@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "user not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "user not found"}`),
		},
		{
			name:  "invalid email format - returns 400",
			email: "invalid-email",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "invalid email format"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "invalid email format"}`),
		},
		{
			name:  "email already verified - returns 409",
			email: "verified@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusConflict,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "email already verified"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "email already verified"}`),
		},
		{
			name:  "rate limit exceeded - returns 429",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusTooManyRequests,
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(`{"error": "rate limit exceeded, try again later"}`),
							),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "rate limit exceeded, try again later"}`),
		},
		{
			name:  "internal server error - returns 500",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "internal server error"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "internal server error"}`),
		},
		{
			name:  "http client error",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
		},
		{
			name:  "empty response body with error status",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(""),
		},
		{
			name:  "email with spaces",
			email: " user@example.com ",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						// Пробелы сохраняются как есть
						assert.Equal(t, " user@example.com ", input.Email)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")
			ctx := context.Background()

			err := repo.SendVerifyEmailMessage(ctx, tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_SendForgetPasswordMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:  "successful send forget password message",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Проверяем метод
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/users/password/forget", req.URL.Path)
						assert.Equal(
							t,
							common.ApplicationJSONContentType,
							req.Header.Get(common.ContentTypeHeaderName),
						)

						// Проверяем тело запроса
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "user@example.com", input.Email)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "send forget password message with special characters",
			email: "user.name+tag@example.co.uk",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "user.name+tag@example.co.uk", input.Email)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "user not found - returns 404",
			email: "nonexistent@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "user not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "user not found"}`),
		},
		{
			name:  "invalid email format - returns 400",
			email: "invalid-email",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "invalid email format"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "invalid email format"}`),
		},
		{
			name:  "rate limit exceeded - returns 429",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusTooManyRequests,
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(`{"error": "rate limit exceeded, try again later"}`),
							),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "rate limit exceeded, try again later"}`),
		},
		{
			name:  "internal server error - returns 500",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "internal server error"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "internal server error"}`),
		},
		{
			name:  "http client error",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
		},
		{
			name:  "empty response body with error status",
			email: "user@example.com",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(""),
		},
		{
			name:  "empty email",
			email: "",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.SendVerifyEmailMessageDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Empty(t, input.Email)

						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body: io.NopCloser(
								bytes.NewReader([]byte(`{"error": "email is required"}`)),
							),
						}, nil
					}).
					Times(1)
			},
			expectedError: errors.New(`{"error": "email is required"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")
			ctx := context.Background()

			err := repo.SendForgetPasswordMessage(ctx, tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_ForgetPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		forgetPasswordToken string
		newPassword         string
		setupMocks          func(*mockhttp.MockHTTPClient)
		expectedError       error
	}{
		{
			name:                "successful forget password",
			forgetPasswordToken: "valid-token-123",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Проверяем метод
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/users/password/forget/valid-token-123", req.URL.Path)
						assert.Equal(
							t,
							common.ApplicationJSONContentType,
							req.Header.Get(common.ContentTypeHeaderName),
						)

						// Проверяем тело запроса
						var input domains.ForgetPasswordDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "NewPassword123!", input.NewPassword)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:                "successful forget password with long password",
			forgetPasswordToken: "valid-token-456",
			newPassword:         "VeryLongPasswordWithManyCharacters123!@#",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.ForgetPasswordDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(
							t,
							"VeryLongPasswordWithManyCharacters123!@#",
							input.NewPassword,
						)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:                "invalid token - returns 400",
			forgetPasswordToken: "invalid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "invalid token"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "invalid token"}`),
		},
		{
			name:                "expired token - returns 401",
			forgetPasswordToken: "expired-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "token has expired"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "token has expired"}`),
		},
		{
			name:                "weak password - returns 400",
			forgetPasswordToken: "valid-token",
			newPassword:         "weak",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(`{"error": "password does not meet security requirements"}`),
							),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "password does not meet security requirements"}`),
		},
		{
			name:                "user not found - returns 404",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "user not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "user not found"}`),
		},
		{
			name:                "internal server error - returns 500",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "internal server error"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`{"error": "internal server error"}`),
		},
		{
			name:                "http client error",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
		},
		{
			name:                "empty token",
			forgetPasswordToken: "",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Проверяем URL с пустым токеном
						assert.Equal(t, "/users/password/forget/", req.URL.Path)

						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body: io.NopCloser(
								bytes.NewReader([]byte(`{"error": "token is required"}`)),
							),
						}, nil
					}).
					Times(1)
			},
			expectedError: errors.New(`{"error": "token is required"}`),
		},
		{
			name:                "empty new password",
			forgetPasswordToken: "valid-token",
			newPassword:         "",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						var input domains.ForgetPasswordDTO

						body, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Empty(t, input.NewPassword)

						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body: io.NopCloser(
								bytes.NewReader([]byte(`{"error": "new password is required"}`)),
							),
						}, nil
					}).
					Times(1)
			},
			expectedError: errors.New(`{"error": "new password is required"}`),
		},
		{
			name:                "empty response body with error status",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")
			ctx := context.Background()

			err := repo.ForgetPassword(ctx, tt.forgetPasswordToken, tt.newPassword)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_ChangePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		accessToken        string
		changePasswordData domains.ChangePasswordDTO
		setupMocks         func(*mockhttp.MockHTTPClient)
		expectedError      error
	}{
		{
			name:        "successful change password",
			accessToken: "valid_access_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/users/password/change", req.URL.Path)
						assert.Equal(
							t,
							common.ApplicationJSONContentType,
							req.Header.Get(common.ContentTypeHeaderName),
						)

						cookie, err := req.Cookie(accessTokenCookieName)
						assert.NoError(t, err)
						assert.Equal(t, "valid_access_token", cookie.Value)

						var input domains.ChangePasswordDTO

						body, _ := io.ReadAll(req.Body)
						err = json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, "OldPassword123!", input.OldPassword)
						assert.Equal(t, "NewPassword123!", input.NewPassword)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:        "wrong old password - returns 400",
			accessToken: "valid_access_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "WrongPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`wrong password`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`wrong password`),
		},
		{
			name:        "user not found - returns 404",
			accessToken: "valid_access_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`user not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`user not found`),
		},
		{
			name:        "unauthorized - returns 401",
			accessToken: "invalid_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`unauthorized`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`unauthorized`),
		},
		{
			name:        "internal server error - returns 500",
			accessToken: "valid_access_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`internal server error`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`internal server error`),
		},
		{
			name:        "http client error",
			accessToken: "valid_access_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
		},
		{
			name:        "empty response body with error status",
			accessToken: "valid_access_token",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte(""))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")
			ctx := context.Background()

			err := repo.ChangePassword(ctx, tt.accessToken, tt.changePasswordData)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
