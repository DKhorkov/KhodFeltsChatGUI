package chats

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
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var validToken = "valid token"

func TestRepository_GetUserChats(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		accessToken   string
		pagination    *domains.Pagination
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedChats []domains.Chat
		expectedError error
	}{
		{
			name:        "successful get user chats",
			accessToken: validToken,
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				chats := []domains.Chat{
					{
						ID:          1,
						Title:       pointers.New("Chat 1"),
						CreatedAt:   now,
						UpdatedAt:   now,
						UnreadCount: 3,
					},
					{
						ID:          2,
						Title:       pointers.New("Chat 2"),
						CreatedAt:   now,
						UpdatedAt:   now,
						UnreadCount: 0,
					},
				}
				chatsData, _ := json.Marshal(chats)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

						query := req.URL.Query()
						if query.Get("limit") != "10" {
							return nil, errors.New("invalid limit parameter")
						}

						if query.Get("offset") != "10" {
							return nil, errors.New("invalid offset parameter")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(chatsData)),
						}, nil
					}).
					Times(1)
			},
			expectedChats: []domains.Chat{
				{
					ID:          1,
					Title:       pointers.New("Chat 1"),
					CreatedAt:   now,
					UpdatedAt:   now,
					UnreadCount: 3,
				},
				{
					ID:          2,
					Title:       pointers.New("Chat 2"),
					CreatedAt:   now,
					UpdatedAt:   now,
					UnreadCount: 0,
				},
			},
			expectedError: nil,
		},
		{
			name:        "empty chats list",
			accessToken: validToken,
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				chatsData, _ := json.Marshal([]domains.Chat{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(chatsData)),
					}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{},
			expectedError: nil,
		},
		{
			name:        "with pagination - large offset",
			accessToken: validToken,
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](20),
				Offset: pointers.New[uint64](100),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				chatsData, _ := json.Marshal([]domains.Chat{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						query := req.URL.Query()
						if query.Get("limit") != "20" {
							return nil, errors.New("invalid limit")
						}

						if query.Get("offset") != "100" {
							return nil, errors.New("invalid offset")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(chatsData)),
						}, nil
					}).
					Times(1)
			},
			expectedChats: []domains.Chat{},
			expectedError: nil,
		},
		{
			name:        "http client error",
			accessToken: "token",
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedChats: nil,
			expectedError: errors.New("connection refused"),
		},
		{
			name:        "unauthorized access",
			accessToken: "invalid_token",
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
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
			expectedChats: nil,
			expectedError: errors.New(`unauthorized`),
		},
		{
			name:        "not found",
			accessToken: "token",
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`chats not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedChats: nil,
			expectedError: errors.New(`chats not found`),
		},
		{
			name:        "invalid json response",
			accessToken: "token",
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
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
			expectedChats: nil,
			expectedError: &json.SyntaxError{},
		},
		{
			name:        "zero limit",
			accessToken: "token",
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				chatsData, _ := json.Marshal([]domains.Chat{})
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(chatsData)),
					}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{},
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

			chats, err := repo.GetUserChats(
				context.Background(),
				tt.accessToken,
				tt.pagination,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			reflect.DeepEqual(tt.expectedChats, chats)
		})
	}
}

func TestRepository_CreateChat(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		accessToken   string
		chat          domains.Chat
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedChat  *domains.Chat
		expectedError error
	}{
		{
			name:        "successful create chat",
			accessToken: validToken,
			chat: domains.Chat{
				Title:     pointers.New("New Chat"),
				Type:      domains.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
				Members:   []domains.User{{ID: 1}},
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				createdChat := domains.Chat{
					ID:        1,
					Title:     pointers.New("New Chat"),
					Type:      domains.ChatTypePrivate,
					CreatedAt: now,
					UpdatedAt: now,
					Members:   []domains.User{{ID: 1}, {ID: 2}},
				}
				createdData, _ := json.Marshal(createdChat)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodPost {
							return nil, errors.New("expected POST method")
						}

						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

						contentType := req.Header.Get("Content-Type")
						if contentType != "application/json" {
							return nil, errors.New("invalid Content-Type")
						}

						var chat domains.Chat

						body, _ := io.ReadAll(req.Body)
						if err := json.Unmarshal(body, &chat); err != nil {
							return nil, err
						}

						if chat.Title != nil && *chat.Title != "New Chat" {
							return nil, errors.New("invalid chat data")
						}

						return &http.Response{
							StatusCode: http.StatusCreated,
							Body:       io.NopCloser(bytes.NewReader(createdData)),
						}, nil
					}).
					Times(1)
			},
			expectedChat: &domains.Chat{
				ID:        1,
				Title:     pointers.New("New Chat"),
				Type:      domains.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
				Members:   []domains.User{{ID: 1}, {ID: 2}},
			},
			expectedError: nil,
		},
		{
			name:        "http client error",
			accessToken: "token",
			chat: domains.Chat{
				Title:     pointers.New("Valid Chat"),
				Type:      domains.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
				Members:   []domains.User{{ID: 1}},
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("network error")).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: errors.New("network error"),
		},
		{
			name:        "conflict - chat already exists",
			accessToken: "token",
			chat: domains.Chat{
				Title:     pointers.New("Existing Chat"),
				Type:      domains.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
				Members:   []domains.User{{ID: 1}},
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusConflict,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "chat already exists"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: internalerrors.ErrCreateChat,
		},
		{
			name:        "bad request - invalid data",
			accessToken: "token",
			chat: domains.Chat{
				Title:     pointers.New("Invalid"),
				Type:      domains.ChatTypePrivate,
				CreatedAt: time.Time{}, // Zero time
				UpdatedAt: time.Time{},
				Members:   []domains.User{{ID: 1}},
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "invalid data"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: internalerrors.ErrCreateChat,
		},
		{
			name:        "invalid json response",
			accessToken: "token",
			chat: domains.Chat{
				Title:     pointers.New("Valid Chat"),
				Type:      domains.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
				Members:   []domains.User{{ID: 1}},
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
			expectedChat:  nil,
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

			createdChat, err := repo.CreateChat(context.Background(), tt.accessToken, tt.chat)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			reflect.DeepEqual(tt.expectedChat, createdChat)
		})
	}
}
