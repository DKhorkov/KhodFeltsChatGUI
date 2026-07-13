package reactions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	mockhttp "github.com/DKhorkov/kfcGUI/mocks/http"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const validToken = "valid token"

func TestRepository_ListReactions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		setupMocks        func(*mockhttp.MockHTTPClient)
		expectedReactions []domains.Reaction
		expectedError     error
	}{
		{
			name: "successful list reactions",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				reactions := []domains.Reaction{
					{ID: 1, Emoji: "👍", SortOrder: 10},
					{ID: 2, Emoji: "❤️", SortOrder: 20},
					{ID: 3, Emoji: "🔥", SortOrder: 30},
				}
				data, _ := json.Marshal(reactions)

				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodGet {
							return nil, errors.New("invalid method")
						}

						if req.URL.Path != "/reactions" {
							return nil, errors.New("invalid URL path: " + req.URL.Path)
						}

						// Публичный роут: cookie быть не должно.
						if _, err := req.Cookie(accessTokenCookieName); err == nil {
							return nil, errors.New("cookie must not be set")
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(data)),
						}, nil
					}).
					Times(1)
			},
			expectedReactions: []domains.Reaction{
				{ID: 1, Emoji: "👍", SortOrder: 10},
				{ID: 2, Emoji: "❤️", SortOrder: 20},
				{ID: 3, Emoji: "🔥", SortOrder: 30},
			},
			expectedError: nil,
		},
		{
			name: "empty dictionary",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`[]`))),
					}, nil).
					Times(1)
			},
			expectedReactions: []domains.Reaction{},
			expectedError:     nil,
		},
		{
			name: "server error",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte(`server error`))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New("server error"),
		},
		{
			name: "http client error",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("timeout")).
					Times(1)
			},
			expectedError: errors.New("timeout"),
		},
		{
			name: "invalid json in response",
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`not-json`))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New("invalid json"),
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

			got, err := repo.ListReactions(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)

				if !reflect.DeepEqual(got, tt.expectedReactions) {
					t.Errorf("got %+v, want %+v", got, tt.expectedReactions)
				}
			}
		})
	}
}

func TestRepository_AddMessageReaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		dto           domains.MessageReactionDTO
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:        "successful add reaction",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodPost {
							return nil, errors.New("invalid method")
						}

						if req.URL.Path != "/messages/10/reactions" {
							return nil, errors.New("invalid URL path: " + req.URL.Path)
						}

						body, err := io.ReadAll(req.Body)
						if err != nil {
							return nil, err
						}

						var dto domains.MessageReactionDTO
						if err = json.Unmarshal(body, &dto); err != nil {
							return nil, err
						}

						if dto.ReactionID != 1 {
							return nil, errors.New("wrong reactionId in body")
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
			name:        "conflict — already exists",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusConflict,
						Body: io.NopCloser(
							bytes.NewReader([]byte(`reaction already exists`)),
						),
					}, nil).
					Times(1)
			},
			expectedError: customerrors.ErrReactionAlreadyExists,
		},
		{
			name:        "reaction not found",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 999,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewReader([]byte(`reaction not found`))),
					}, nil).
					Times(1)
			},
			expectedError: customerrors.ErrReactionNotFound,
		},
		{
			name:        "forbidden — not chat member",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusForbidden,
						Body:       io.NopCloser(bytes.NewReader([]byte(`not a chat member`))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New("not a chat member"),
		},
		{
			name:        "http client error",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("network")).
					Times(1)
			},
			expectedError: errors.New("network"),
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

			err := repo.AddMessageReaction(context.Background(), tt.accessToken, tt.dto)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_RemoveMessageReaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accessToken   string
		dto           domains.MessageReactionDTO
		setupMocks    func(*mockhttp.MockHTTPClient)
		expectedError error
	}{
		{
			name:        "successful remove reaction",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						if req.Method != http.MethodDelete {
							return nil, errors.New("invalid method")
						}

						if req.URL.Path != "/messages/10/reactions/1" {
							return nil, errors.New("invalid URL path: " + req.URL.Path)
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
			name:        "message not found",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  999,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewReader([]byte(`message not found`))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New("message not found"),
		},
		{
			name:        "forbidden",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusForbidden,
						Body:       io.NopCloser(bytes.NewReader([]byte(`forbidden`))),
					}, nil).
					Times(1)
			},
			expectedError: errors.New("forbidden"),
		},
		{
			name:        "http client error",
			accessToken: validToken,
			dto: domains.MessageReactionDTO{
				MessageID:  10,
				ReactionID: 1,
			},
			setupMocks: func(mockClient *mockhttp.MockHTTPClient) {
				mockClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, errors.New("network")).
					Times(1)
			},
			expectedError: errors.New("network"),
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

			err := repo.RemoveMessageReaction(context.Background(), tt.accessToken, tt.dto)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
