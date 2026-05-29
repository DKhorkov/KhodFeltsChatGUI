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
	"github.com/DKhorkov/libs/pointers"
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

func TestRepository_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		accessToken    string
		updateUserData domains.UpdateUserDTO
		setupMocks     func(*mockhttp.MockHTTPClient)
		expectedUser   *domains.User
		expectedError  error
	}{
		{
			name:           "successful update user",
			accessToken:    "valid_access_token",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				user := domains.User{
					ID:       1,
					Username: "newusername",
					Email:    "john@example.com",
				}
				userData, _ := json.Marshal(user)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPut, req.Method)
						assert.Equal(t, "/users/me", req.URL.Path)
						assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

						cookie, err := req.Cookie(accessTokenCookieName)
						assert.NoError(t, err)
						assert.Equal(t, "valid_access_token", cookie.Value)

						var input domains.UpdateUserDTO

						body, _ := io.ReadAll(req.Body)
						err = json.Unmarshal(body, &input)
						assert.NoError(t, err)
						assert.Equal(t, pointers.New("newusername"), input.Username)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(userData)),
						}, nil
					}).
					Times(1)
			},
			expectedUser: &domains.User{
				ID:       1,
				Username: "newusername",
				Email:    "john@example.com",
			},
			expectedError: nil,
		},
		{
			name:           "validation error - returns 400",
			accessToken:    "valid_access_token",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("ab")},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`validation failed`)),
						),
					}, nil).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New(`validation failed`),
		},
		{
			name:           "user not found - returns 404",
			accessToken:    "valid_access_token",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
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
			name:           "unauthorized - returns 401",
			accessToken:    "invalid_token",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
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
			name:           "internal server error - returns 500",
			accessToken:    "valid_access_token",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
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
			expectedUser:  nil,
			expectedError: errors.New(`internal server error`),
		},
		{
			name:           "http client error",
			accessToken:    "valid_access_token",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
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
			name:           "invalid json response",
			accessToken:    validToken,
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
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

			user, err := repo.UpdateUser(context.Background(), tt.accessToken, tt.updateUserData)

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
		filters       *domains.UsersFilters
		pagination    *domains.Pagination
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedUsers []domains.User
		expectedError error
	}{
		{
			name: "successful search users",
			filters: &domains.UsersFilters{
				Username: pointers.New("john"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
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

						if query.Get("offset") != "10" {
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
			name: "search with special characters in username",
			filters: &domains.UsersFilters{
				Username: pointers.New("john@domain"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
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
			name: "empty search results",
			filters: &domains.UsersFilters{
				Username: pointers.New("nonexistent"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
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
			name: "search with zero limit",
			filters: &domains.UsersFilters{
				Username: pointers.New("test"),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				usersData, _ := json.Marshal([]domains.User{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(_ *http.Request) (*http.Response, error) {
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
			name: "http client error",
			filters: &domains.UsersFilters{
				Username: pointers.New("john"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
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
			name: "user not found status",
			filters: &domains.UsersFilters{
				Username: pointers.New("unknown"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
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
			name: "invalid json response",
			filters: &domains.UsersFilters{
				Username: pointers.New("john"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
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
			name: "large limit and offset",
			filters: &domains.UsersFilters{
				Username: pointers.New("test"),
			},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](100000),
				Offset: pointers.New[uint64](5000000),
			},
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

			users, err := repo.SearchUsers(context.Background(), tt.filters, tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUsers, users)
		})
	}
}

func TestRepository_UpdateAvatar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		fileData      []byte
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedURL   string
		expectedError error
	}{
		{
			name:        "successful avatar upload",
			accessToken: "valid_access_token",
			fileData:    []byte("fake image data"),
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPut, req.Method)
						assert.Equal(t, "/users/me/avatar", req.URL.Path)
						assert.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data")

						cookie, err := req.Cookie(accessTokenCookieName)
						assert.NoError(t, err)
						assert.Equal(t, "valid_access_token", cookie.Value)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader([]byte("https://kfc.webtm.ru/api/files/download/uuid.jpg"))),
						}, nil
					}).
					Times(1)
			},
			expectedURL:   "https://kfc.webtm.ru/api/files/download/uuid.jpg",
			expectedError: nil,
		},
		{
			name:        "bad request - invalid format",
			accessToken: "valid_access_token",
			fileData:    []byte("not an image"),
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte("invalid image format"))),
					}, nil).
					Times(1)
			},
			expectedURL:   "",
			expectedError: errors.New("invalid image format"),
		},
		{
			name:        "http client error",
			accessToken: "valid_access_token",
			fileData:    []byte("data"),
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedURL:   "",
			expectedError: errors.New("connection refused"),
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

			avatarURL, err := repo.UpdateAvatar(context.Background(), tt.accessToken, tt.fileData)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedURL, avatarURL)
		})
	}
}

func TestRepository_DeleteAvatar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:        "successful avatar delete",
			accessToken: "valid_access_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodDelete, req.Method)
						assert.Equal(t, "/users/me/avatar", req.URL.Path)

						cookie, err := req.Cookie(accessTokenCookieName)
						assert.NoError(t, err)
						assert.Equal(t, "valid_access_token", cookie.Value)

						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(bytes.NewReader(nil)),
						}, nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:        "server error",
			accessToken: "valid_access_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte("internal server error"))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New("internal server error"),
		},
		{
			name:        "http client error",
			accessToken: "valid_access_token",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
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

			err := repo.DeleteAvatar(context.Background(), tt.accessToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
