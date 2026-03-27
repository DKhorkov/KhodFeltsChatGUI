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
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestChatsRepository_GetUserChats(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		accessToken   string
		limit         int
		offset        int
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedChats []domains.Chat
		expectedError error
	}{
		{
			name:        "successful get user chats",
			accessToken: "valid_token",
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				chats := []domains.Chat{
					{
						ID:        1,
						Title:     pointers.New("Chat 1"),
						CreatedAt: now,
						UpdatedAt: now,
					},
					{
						ID:        2,
						Title:     pointers.New("Chat 2"),
						CreatedAt: now,
						UpdatedAt: now,
					},
				}
				chatsData, _ := json.Marshal(chats)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != "valid_token" {
							return nil, errors.New("invalid cookie")
						}

						query := req.URL.Query()
						if query.Get("limit") != "10" {
							return nil, errors.New("invalid limit parameter")
						}

						if query.Get("offset") != "0" {
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
					ID:        1,
					Title:     pointers.New("Chat 1"),
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        2,
					Title:     pointers.New("Chat 2"),
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
		{
			name:        "empty chats list",
			accessToken: "valid_token",
			limit:       10,
			offset:      0,
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
			accessToken: "valid_token",
			limit:       20,
			offset:      100,
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
			limit:       10,
			offset:      0,
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
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "unauthorized"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedChats: nil,
			expectedError: internalerrors.ErrGetUserChats,
		},
		{
			name:        "not found",
			accessToken: "token",
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "chats not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedChats: nil,
			expectedError: internalerrors.ErrGetUserChats,
		},
		{
			name:        "invalid json response",
			accessToken: "token",
			limit:       10,
			offset:      0,
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
			limit:       0,
			offset:      0,
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
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewChatsRepository(mockClient, "http://api.example.com")

			chats, err := repo.GetUserChats(
				context.Background(),
				tt.accessToken,
				tt.limit,
				tt.offset,
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

func TestChatsRepository_CreateChat(t *testing.T) {
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
			accessToken: "valid_token",
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
						if err != nil || cookie.Value != "valid_token" {
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
			name:        "invalid chat - empty name",
			accessToken: "token",
			chat: domains.Chat{
				Title:     pointers.New(""),
				CreatedAt: now,
				UpdatedAt: now,
			},
			setupMocks:    func(mockClient *mockhttp.MockHTTPClient) {},
			expectedChat:  nil,
			expectedError: internalerrors.ErrCreateChat,
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
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewChatsRepository(mockClient, "http://api.example.com")

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

func TestChatsRepository_GetChatMessages(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		accessToken      string
		chatID           uint64
		limit            int
		offset           int
		setupMocks       func(*mockhttp.MockHTTPClient)
		expectedMessages []domains.Message
		expectedError    error
	}{
		{
			name:        "successful get messages",
			accessToken: "valid_token",
			chatID:      1,
			limit:       20,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				messages := []domains.Message{
					{
						ID:     1,
						ChatID: 1,
						Sender: domains.User{
							ID:             100,
							Username:       "john_doe",
							Email:          "john@example.com",
							EmailConfirmed: true,
							Password:       "",
							CreatedAt:      now,
							UpdatedAt:      now,
						},
						Text:      "Hello!",
						CreatedAt: now,
						UpdatedAt: now,
					},
					{
						ID:     2,
						ChatID: 1,
						Sender: domains.User{
							ID:             101,
							Username:       "jane_doe",
							Email:          "jane@example.com",
							EmailConfirmed: true,
							Password:       "",
							CreatedAt:      now,
							UpdatedAt:      now,
						},
						Text:      "Hi there!",
						CreatedAt: now,
						UpdatedAt: now,
					},
				}
				messagesData, _ := json.Marshal(messages)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						expectedURL := "/chats/1/messages"
						if req.URL.Path != expectedURL {
							return nil, errors.New("invalid URL path")
						}

						query := req.URL.Query()
						if query.Get("limit") != "20" {
							return nil, errors.New("invalid limit")
						}

						if query.Get("offset") != "0" {
							return nil, errors.New("invalid offset")
						}

						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != "valid_token" {
							return nil, errors.New("invalid cookie")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(messagesData)),
						}, nil
					}).
					Times(1)
			},
			expectedMessages: []domains.Message{
				{
					ID:     1,
					ChatID: 1,
					Sender: domains.User{
						ID:             100,
						Username:       "john_doe",
						Email:          "john@example.com",
						EmailConfirmed: true,
						Password:       "",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
					Text:      "Hello!",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:     2,
					ChatID: 1,
					Sender: domains.User{
						ID:             101,
						Username:       "jane_doe",
						Email:          "jane@example.com",
						EmailConfirmed: true,
						Password:       "",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
					Text:      "Hi there!",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
		{
			name:        "empty messages list",
			accessToken: "token",
			chatID:      1,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				messagesData, _ := json.Marshal([]domains.Message{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(messagesData)),
					}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{},
			expectedError:    nil,
		},
		{
			name:        "message with empty sender",
			accessToken: "token",
			chatID:      2,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				messages := []domains.Message{
					{
						ID:        10,
						ChatID:    2,
						Sender:    domains.User{},
						Text:      "System message",
						CreatedAt: now,
						UpdatedAt: now,
					},
				}
				messagesData, _ := json.Marshal(messages)

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(messagesData)),
					}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{
				{
					ID:        10,
					ChatID:    2,
					Sender:    domains.User{},
					Text:      "System message",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
		{
			name:        "with pagination",
			accessToken: "token",
			chatID:      5,
			limit:       50,
			offset:      100,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				messagesData, _ := json.Marshal([]domains.Message{})

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						query := req.URL.Query()
						if query.Get("limit") != "50" {
							return nil, errors.New("invalid limit")
						}

						if query.Get("offset") != "100" {
							return nil, errors.New("invalid offset")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(messagesData)),
						}, nil
					}).
					Times(1)
			},
			expectedMessages: []domains.Message{},
			expectedError:    nil,
		},
		{
			name:        "chat not found",
			accessToken: "token",
			chatID:      999,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "chat not found"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    internalerrors.ErrGetChatMessages,
		},
		{
			name:        "unauthorized access",
			accessToken: "invalid_token",
			chatID:      1,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusUnauthorized,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "unauthorized"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    internalerrors.ErrGetChatMessages,
		},
		{
			name:        "http client error",
			accessToken: "token",
			chatID:      1,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("timeout")).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New("timeout"),
		},
		{
			name:        "invalid json response",
			accessToken: "token",
			chatID:      1,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{invalid json}`))),
					}, nil).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    &json.SyntaxError{},
		},
		{
			name:        "forbidden access",
			accessToken: "token",
			chatID:      1,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusForbidden,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`{"error": "access denied"}`)),
						),
					}, nil).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    internalerrors.ErrGetChatMessages,
		},
		{
			name:        "message with all fields populated",
			accessToken: "token",
			chatID:      3,
			limit:       10,
			offset:      0,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				messages := []domains.Message{
					{
						ID:     100,
						ChatID: 3,
						Sender: domains.User{
							ID:             200,
							Username:       "test_user",
							Email:          "test@example.com",
							EmailConfirmed: true,
							Password:       "hashed_password",
							CreatedAt:      now,
							UpdatedAt:      now,
						},
						Text:      "Test message with all fields",
						CreatedAt: now,
						UpdatedAt: now,
					},
				}
				messagesData, _ := json.Marshal(messages)

				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(messagesData)),
					}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{
				{
					ID:     100,
					ChatID: 3,
					Sender: domains.User{
						ID:             200,
						Username:       "test_user",
						Email:          "test@example.com",
						EmailConfirmed: true,
						Password:       "hashed_password",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
					Text:      "Test message with all fields",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := NewChatsRepository(mockClient, "http://api.example.com")

			messages, err := repo.GetChatMessages(
				context.Background(),
				tt.accessToken,
				tt.chatID,
				tt.limit,
				tt.offset,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			reflect.DeepEqual(tt.expectedMessages, messages)
		})
	}
}
