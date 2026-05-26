package messages

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
	mockhttp "github.com/DKhorkov/kfcGUI/mocks/http"
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var validToken = "valid token"

func TestRepository_GetChatMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name             string
		accessToken      string
		chatID           uint64
		pagination       *domains.Pagination
		setupMocks       func(*mockhttp.MockHTTPClient)
		expectedMessages []domains.Message
		expectedError    error
	}{
		{
			name:        "successful get messages",
			accessToken: validToken,
			chatID:      1,
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](20),
				Offset: pointers.New[uint64](10),
			},
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

						if query.Get("offset") != "10" {
							return nil, errors.New("invalid offset")
						}

						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
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
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
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
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
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
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](50),
				Offset: pointers.New[uint64](100),
			},
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
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`chat not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New(`chat not found`),
		},
		{
			name:        "unauthorized access",
			accessToken: "invalid_token",
			chatID:      1,
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
			expectedMessages: nil,
			expectedError:    errors.New(`unauthorized`),
		},
		{
			name:        "http client error",
			accessToken: "token",
			chatID:      1,
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
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
			expectedMessages: nil,
			expectedError:    &json.SyntaxError{},
		},
		{
			name:        "forbidden access",
			accessToken: "token",
			chatID:      1,
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusForbidden,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`access denied`)),
						),
					}, nil).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New(`access denied`),
		},
		{
			name:        "message with all fields populated",
			accessToken: "token",
			chatID:      3,
			pagination: &domains.Pagination{
				Limit: pointers.New[uint64](10),
			},
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
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockClient := mockhttp.NewMockHTTPClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockClient)
			}

			repo := New(mockClient, "http://api.example.com")

			messages, err := repo.GetChatMessages(
				context.Background(),
				tt.accessToken,
				tt.chatID,
				tt.pagination,
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

func TestRepository_GetMessageByID(t *testing.T) {
	t.Parallel()

	now := time.Now().Round(0) // Strip monotonic clock reading for JSON round-trip comparison

	tests := []struct {
		name            string
		accessToken     string
		messageID       uint64
		setupMocks      func(*mockhttp.MockHTTPClient)
		expectedMessage *domains.Message
		expectedError   error
	}{
		{
			name:        "successful get message",
			accessToken: validToken,
			messageID:   42,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				message := domains.Message{
					ID:     42,
					ChatID: 1,
					Sender: domains.User{
						ID:             100,
						Username:       "john_doe",
						Email:          "john@example.com",
						EmailConfirmed: true,
						CreatedAt:      now,
						UpdatedAt:      now,
					},
					Text:      "Hello!",
					CreatedAt: now,
					UpdatedAt: now,
				}
				messageData, _ := json.Marshal(message)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						expectedURL := "/messages/42"
						if req.URL.Path != expectedURL {
							return nil, errors.New("invalid URL path")
						}

						if req.Method != http.MethodGet {
							return nil, errors.New("invalid method")
						}

						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(messageData)),
						}, nil
					}).
					Times(1)
			},
			expectedMessage: &domains.Message{
				ID:     42,
				ChatID: 1,
				Sender: domains.User{
					ID:             100,
					Username:       "john_doe",
					Email:          "john@example.com",
					EmailConfirmed: true,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				Text:      "Hello!",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name:        "message not found",
			accessToken: validToken,
			messageID:   999,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`message not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New(`message not found`),
		},
		{
			name:        "unauthorized",
			accessToken: "invalid_token",
			messageID:   42,
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
			expectedMessage: nil,
			expectedError:   errors.New(`unauthorized`),
		},
		{
			name:        "http client error",
			accessToken: validToken,
			messageID:   42,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("timeout")).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("timeout"),
		},
		{
			name:        "invalid json response",
			accessToken: validToken,
			messageID:   42,
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{invalid json}`))),
					}, nil).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   &json.SyntaxError{},
		},
		{
			name:        "internal server error",
			accessToken: validToken,
			messageID:   42,
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
			expectedMessage: nil,
			expectedError:   errors.New(`internal server error`),
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

			message, err := repo.GetMessageByID(
				context.Background(),
				tt.accessToken,
				tt.messageID,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, message)
			}
		})
	}
}

func TestRepository_DeleteMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		dto           domains.DeleteMessageDTO
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:        "successful delete for self",
			accessToken: validToken,
			dto: domains.DeleteMessageDTO{
				MessageID: 1,
				ForAll:    false,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodDelete {
							return nil, errors.New("invalid method")
						}

						expectedURL := "/messages/1"
						if req.URL.Path != expectedURL {
							return nil, errors.New("invalid URL path")
						}

						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

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
			name:        "successful delete for all",
			accessToken: validToken,
			dto: domains.DeleteMessageDTO{
				MessageID: 2,
				ForAll:    true,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodDelete {
							return nil, errors.New("invalid method")
						}

						expectedURL := "/messages/2"
						if req.URL.Path != expectedURL {
							return nil, errors.New("invalid URL path")
						}

						body, err := io.ReadAll(req.Body)
						if err != nil {
							return nil, err
						}

						var dto domains.DeleteMessageDTO
						if err = json.Unmarshal(body, &dto); err != nil {
							return nil, err
						}

						if !dto.ForAll {
							return nil, errors.New("expected forAll to be true")
						}

						cookie, err := req.Cookie(accessTokenCookieName)
						if err != nil || cookie.Value != validToken {
							return nil, errors.New("invalid cookie")
						}

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
			name:        "unauthorized",
			accessToken: "invalid_token",
			dto: domains.DeleteMessageDTO{
				MessageID: 1,
				ForAll:    false,
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
			name:        "message not found",
			accessToken: validToken,
			dto: domains.DeleteMessageDTO{
				MessageID: 999,
				ForAll:    false,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`message not found`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`message not found`),
		},
		{
			name:        "forbidden",
			accessToken: validToken,
			dto: domains.DeleteMessageDTO{
				MessageID: 1,
				ForAll:    false,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusForbidden,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`access denied`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: errors.New(`access denied`),
		},
		{
			name:        "http client error",
			accessToken: validToken,
			dto: domains.DeleteMessageDTO{
				MessageID: 1,
				ForAll:    false,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("timeout")).
					Times(1)
			},
			expectedError: errors.New("timeout"),
		},
		{
			name:        "internal server error",
			accessToken: validToken,
			dto: domains.DeleteMessageDTO{
				MessageID: 1,
				ForAll:    false,
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

			err := repo.DeleteMessage(
				context.Background(),
				tt.accessToken,
				tt.dto,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
