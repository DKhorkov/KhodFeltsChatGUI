package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	mockhttp "github.com/DKhorkov/kfcGUI/mocks/http"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	testBaseURL = "http://localhost:8080/api"
	validToken  = "valid token"
)

func TestRepository_GetSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		accessToken      string
		setupMocks       func(*mockhttp.MockHTTPClient)
		expectedSettings *domains.Settings
		expectedError    error
	}{
		{
			name:        "successful get settings with light theme",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				settings := domains.Settings{
					Theme: domains.ThemeLight,
				}
				data, _ := json.Marshal(settings)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

						if req.Method != http.MethodGet {
							return nil, errors.New("invalid method")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(data)),
						}, nil
					}).
					Times(1)
			},
			expectedSettings: &domains.Settings{
				Theme: domains.ThemeLight,
			},
			expectedError: nil,
		},
		{
			name:        "successful get settings with dark theme",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				settings := domains.Settings{
					Theme: domains.ThemeDark,
				}
				data, _ := json.Marshal(settings)

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(data)),
					}, nil).
					Times(1)
			},
			expectedSettings: &domains.Settings{
				Theme: domains.ThemeDark,
			},
			expectedError: nil,
		},
		{
			name:        "unauthorized access",
			accessToken: "invalid_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body:       io.NopCloser(bytes.NewReader([]byte("unauthorized"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("unauthorized"),
		},
		{
			name:        "settings not found",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewReader([]byte("settings not found"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("settings not found"),
		},
		{
			name:        "http client error",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("connection refused"),
		},
		{
			name:        "server internal error",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte("internal server error"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("internal server error"),
		},
		{
			name:        "invalid json response",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("{invalid json}"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("invalid character"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			tt.setupMocks(mockClient)

			repo := New(mockClient, testBaseURL)
			result, err := repo.GetSettings(context.Background(), tt.accessToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSettings, result)
		})
	}
}

func TestRepository_UpdateSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		accessToken      string
		settings         domains.Settings
		setupMocks       func(*mockhttp.MockHTTPClient)
		expectedSettings *domains.Settings
		expectedError    error
	}{
		{
			name:        "successful update to dark theme",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				responseSettings := domains.Settings{
					Theme: domains.ThemeDark,
				}
				data, _ := json.Marshal(responseSettings)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

						if req.Method != http.MethodPut {
							return nil, errors.New("invalid method")
						}

						if req.Header.Get("Content-Type") != "application/json" {
							return nil, errors.New("invalid content type")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(data)),
						}, nil
					}).
					Times(1)
			},
			expectedSettings: &domains.Settings{
				Theme: domains.ThemeDark,
			},
			expectedError: nil,
		},
		{
			name:        "successful update to light theme",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeLight,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				responseSettings := domains.Settings{
					Theme: domains.ThemeLight,
				}
				data, _ := json.Marshal(responseSettings)

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(data)),
					}, nil).
					Times(1)
			},
			expectedSettings: &domains.Settings{
				Theme: domains.ThemeLight,
			},
			expectedError: nil,
		},
		{
			name:        "unauthorized access",
			accessToken: "invalid_token",
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body:       io.NopCloser(bytes.NewReader([]byte("unauthorized"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("unauthorized"),
		},
		{
			name:        "settings not found",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewReader([]byte("settings not found"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("settings not found"),
		},
		{
			name:        "http client error",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("connection refused"),
		},
		{
			name:        "bad request",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte("bad request"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("bad request"),
		},
		{
			name:        "server internal error",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte("internal server error"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("internal server error"),
		},
		{
			name:        "invalid json response",
			accessToken: validToken,
			settings: domains.Settings{
				Theme: domains.ThemeDark,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("{invalid json}"))),
					}, nil).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("invalid character"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			tt.setupMocks(mockClient)

			repo := New(mockClient, testBaseURL)
			result, err := repo.UpdateSettings(context.Background(), tt.accessToken, tt.settings)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSettings, result)
		})
	}
}
