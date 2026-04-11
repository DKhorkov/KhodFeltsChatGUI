package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	mockhttp "github.com/DKhorkov/kfcGUI/mocks/http"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var validToken = "valid token"

func TestRepository_GetCurrentUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name:        "successful get current user",
			accessToken: "valid_token_123",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				user := domains.User{
					ID:       1,
					Username: "john_doe",
					Email:    "john@example.com",
				}
				userData, _ := json.Marshal(user)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Проверяем cookie
						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != "valid_token_123" {
							return nil, errors.New("invalid cookie")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(userData)),
						}, nil
					}).
					Times(1)
			},
			expectedUser: &domains.User{
				ID:       1,
				Username: "john_doe",
				Email:    "john@example.com",
			},
			expectedError: nil,
		},
		{
			name:        "user not found",
			accessToken: "invalid_token",
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
			expectedUser:  nil,
			expectedError: errors.New(`user not found`),
		},
		{
			name:        "http client error",
			accessToken: "some_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("network error")).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("network error"),
		},
		{
			name:        "invalid json response",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{invalid json}`))),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: &json.SyntaxError{},
		},
		{
			name:        "unauthorized status code",
			accessToken: "expired_token",
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
			expectedUser:  nil,
			expectedError: errors.New(`unauthorized`),
		},
		{
			name:        "internal server error",
			accessToken: "some_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`server error`)),
						),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New(`server error`),
		},
		{
			name:        "empty response body",
			accessToken: validToken,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
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

			baseURL := "http://api.example.com"
			if tt.name == "request creation fails - empty base URL" {
				baseURL = ""
			}

			repo := New(mockClient, baseURL)

			user, err := repo.GetCurrentUser(context.Background(), tt.accessToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)
		})
	}
}

func TestRepository_SearchUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		username      string
		limit         int
		offset        int
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedUsers []domains.User
		expectedError error
	}{
		{
			name:     "successful search users",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				users := []domains.User{
					{ID: 1, Username: "john_doe", Email: "john@example.com"},
					{ID: 2, Username: "john_smith", Email: "john.smith@example.com"},
				}
				usersData, _ := json.Marshal(users)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Проверяем query параметры
						query := req.URL.Query()
						if query.Get("username") != "john" {
							return nil, fmt.Errorf(
								"expected username=john, got %s",
								query.Get("username"),
							)
						}

						if query.Get("limit") != "10" {
							return nil, fmt.Errorf("expected limit=10, got %s", query.Get("limit"))
						}

						if query.Get("offset") != "0" {
							return nil, fmt.Errorf("expected offset=0, got %s", query.Get("offset"))
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(usersData)),
						}, nil
					}).
					Times(1)
			},
			expectedUsers: []domains.User{
				{ID: 1, Username: "john_doe", Email: "john@example.com"},
				{ID: 2, Username: "john_smith", Email: "john.smith@example.com"},
			},
			expectedError: nil,
		},
		{
			name:     "search with special characters in username",
			username: "john@domain",
			limit:    5,
			offset:   10,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				users := []domains.User{}
				usersData, _ := json.Marshal(users)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						query := req.URL.Query()
						// Проверяем, что username закодирован
						if query.Get("username") != "john@domain" {
							return nil, fmt.Errorf(
								"expected username=john@domain, got %s",
								query.Get("username"),
							)
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(usersData)),
						}, nil
					}).
					Times(1)
			},
			expectedUsers: []domains.User{},
			expectedError: nil,
		},
		{
			name:     "empty search results",
			username: "nonexistent",
			limit:    10,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				usersData, _ := json.Marshal([]domains.User{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(usersData)),
					}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{},
			expectedError: nil,
		},
		{
			name:     "search with zero limit",
			username: "test",
			limit:    0,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				usersData, _ := json.Marshal([]domains.User{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						query := req.URL.Query()
						if query.Get("limit") != "0" {
							return nil, fmt.Errorf("expected limit=0, got %s", query.Get("limit"))
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(usersData)),
						}, nil
					}).
					Times(1)
			},
			expectedUsers: []domains.User{},
			expectedError: nil,
		},
		{
			name:     "http client error",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("connection refused"),
		},
		{
			name:     "user not found status",
			username: "unknown",
			limit:    10,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`users not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New(`users not found`),
		},
		{
			name:     "bad request",
			username: "john",
			limit:    -1,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`invalid parameters`)),
						),
					}, nil).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New(`invalid parameters`),
		},
		{
			name:     "invalid json response",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{invalid json}`))),
					}, nil).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: &json.SyntaxError{},
		},
		{
			name:     "large limit and offset",
			username: "test",
			limit:    1000,
			offset:   5000,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				usersData, _ := json.Marshal([]domains.User{})
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(usersData)),
					}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{},
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

			users, err := repo.SearchUsers(context.Background(), tt.username, tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUsers, users)
		})
	}
}
