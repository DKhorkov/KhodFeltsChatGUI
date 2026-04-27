package chat_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	chathandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/chat"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	loggingmocks "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful get current user",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1, Username: "john"}, nil).
					Times(1)
			},
			expectedUser:  &domains.User{ID: 1, Username: "john"},
			expectedError: nil,
		},
		{
			name: "use case error",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(nil, errors.New("session expired")).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("session expired"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil, config.ValidationConfig{})

			user, err := h.GetCurrentUser()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestHandler_GetUserChats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pagination    *domains.Pagination
		setupMocks    func(*mockusecases.MockUseCases)
		expectedChats []domains.Chat
		expectedError error
	}{
		{
			name:       "successful get chats without pagination",
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), (*domains.Pagination)(nil)).
					Return([]domains.Chat{{ID: 1}, {ID: 2}}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{{ID: 1}, {ID: 2}},
			expectedError: nil,
		},
		{
			name: "successful get chats with pagination",
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](0),
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), &domains.Pagination{Limit: pointers.New[uint64](10), Offset: pointers.New[uint64](0)}).
					Return([]domains.Chat{{ID: 1}}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{{ID: 1}},
			expectedError: nil,
		},
		{
			name:       "empty chats list",
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), gomock.Any()).
					Return([]domains.Chat{}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{},
			expectedError: nil,
		},
		{
			name:       "use case error",
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), gomock.Any()).
					Return(nil, customerrors.ErrGetUserChats).
					Times(1)
			},
			expectedChats: nil,
			expectedError: customerrors.ErrGetUserChats,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil, config.ValidationConfig{})

			chats, err := h.GetUserChats(tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, chats)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedChats, chats)
			}
		})
	}
}

func TestHandler_GetChatMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		chatID           uint64
		pagination       *domains.Pagination
		setupMocks       func(*mockusecases.MockUseCases)
		expectedMessages []domains.Message
		expectedError    error
	}{
		{
			name:   "successful get messages",
			chatID: 42,
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](0),
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(42), &domains.Pagination{Limit: pointers.New[uint64](10), Offset: pointers.New[uint64](0)}).
					Return([]domains.Message{{ID: 1, Text: "hello"}, {ID: 2, Text: "world"}}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{{ID: 1, Text: "hello"}, {ID: 2, Text: "world"}},
			expectedError:    nil,
		},
		{
			name:       "get messages without pagination",
			chatID:     42,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(42), (*domains.Pagination)(nil)).
					Return([]domains.Message{}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{},
			expectedError:    nil,
		},
		{
			name:       "use case error",
			chatID:     42,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(42), gomock.Any()).
					Return(nil, customerrors.ErrGetChatMessages).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    customerrors.ErrGetChatMessages,
		},
		{
			name:       "chat not found",
			chatID:     999,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(999), gomock.Any()).
					Return(nil, customerrors.ErrChatNotFound).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    customerrors.ErrChatNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil, config.ValidationConfig{})

			messages, err := h.GetChatMessages(tt.chatID, tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, messages)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessages, messages)
			}
		})
	}
}

func TestHandler_SendMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		chatID        uint64
		text          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedError error
	}{
		{
			name:   "successful send message",
			chatID: 42,
			text:   "Hello!",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1, Username: "john"}, nil).
					Times(1)
				uc.EXPECT().
					SendMessage(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:   "get current user error",
			chatID: 42,
			text:   "Hello!",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(nil, errors.New("session expired")).
					Times(1)
			},
			expectedError: errors.New("session expired"),
		},
		{
			name:   "send message use case error",
			chatID: 42,
			text:   "Hello!",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1}, nil).
					Times(1)
				uc.EXPECT().
					SendMessage(gomock.Any(), gomock.Any()).
					Return(errors.New("websocket closed")).
					Times(1)
			},
			expectedError: errors.New("websocket closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil, config.ValidationConfig{})

			err := h.SendMessage(tt.chatID, tt.text)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_ToggleTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedTheme domains.ThemeType
		expectedError error
	}{
		{
			name: "toggle from light to dark",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeLight).
					Times(1)
				uc.EXPECT().
					SetTheme(gomock.Any(), domains.ThemeDark).
					Return(nil).
					Times(1)
			},
			expectedTheme: domains.ThemeDark,
			expectedError: nil,
		},
		{
			name: "toggle from dark to light",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeDark).
					Times(1)
				uc.EXPECT().
					SetTheme(gomock.Any(), domains.ThemeLight).
					Return(nil).
					Times(1)
			},
			expectedTheme: domains.ThemeLight,
			expectedError: nil,
		},
		{
			name: "set theme error returns ThemeLight",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeLight).
					Times(1)
				uc.EXPECT().
					SetTheme(gomock.Any(), domains.ThemeDark).
					Return(errors.New("storage error")).
					Times(1)
			},
			expectedTheme: domains.ThemeLight,
			expectedError: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil, config.ValidationConfig{})

			theme, err := h.ToggleTheme()

			assert.Equal(t, tt.expectedTheme, theme)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := chathandler.New(mockUseCases, mockMapper, nil, config.ValidationConfig{})
	h.SetContext(context.Background())
}

func TestHandler_StartListening_StopListening(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(*mockusecases.MockUseCases, *loggingmocks.MockLogger)
	}{
		{
			name: "start and stop with ReadMessage returning ErrWebsocketClosed",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				// readMessages goroutine: exits on ErrWebsocketClosed
				uc.EXPECT().
					ReadMessage(gomock.Any()).
					Return((*domains.Message)(nil), customerrors.ErrWebsocketClosed).
					AnyTimes()
				// logging.LogInfo is called inside readMessages on ErrWebsocketClosed
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				// logging.LogErrorContext may also be called
				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name:       "stop without start does not panic",
			setupMocks: func(_ *mockusecases.MockUseCases, _ *loggingmocks.MockLogger) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)
			mockLogger := loggingmocks.NewMockLogger(ctrl)

			tt.setupMocks(mockUseCases, mockLogger)

			h := chathandler.New(mockUseCases, mockMapper, mockLogger, config.ValidationConfig{})

			if tt.name != "stop without start does not panic" {
				h.StartListening()
			}

			h.StopListening()
		})
	}
}
