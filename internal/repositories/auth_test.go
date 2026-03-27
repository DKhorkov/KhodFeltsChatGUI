package repositories

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

	"github.com/DKhorkov/kfcGUI/internal/domains"
	internalerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	mockhttp "github.com/DKhorkov/kfcGUI/mocks/http"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthRepository_Register(t *testing.T) {
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
							bytes.NewReader([]byte(`{"error": "email already exists"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: internalerrors.ErrRegister,
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
							bytes.NewReader([]byte(`{"error": "invalid email format"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: internalerrors.ErrRegister,
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
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewAuthRepository(mockClient, "http://api.example.com")

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

func TestAuthRepository_Login(t *testing.T) {
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
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewAuthRepository(mockClient, "http://api.example.com")

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

func TestAuthRepository_Logout(t *testing.T) {
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
							bytes.NewReader([]byte(`{"error": "invalid token"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: internalerrors.ErrLogout,
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
							bytes.NewReader([]byte(`{"error": "session not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: internalerrors.ErrLogout,
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
			expectedError: internalerrors.ErrLogout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewAuthRepository(mockClient, "http://api.example.com")

			err := repo.Logout(context.Background(), tt.accessToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_RefreshTokens(t *testing.T) {
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
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewAuthRepository(mockClient, "http://api.example.com")

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
