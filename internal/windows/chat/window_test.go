package chat

import (
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	mockwindows "github.com/DKhorkov/kfcGUI/mocks/window"
	"github.com/DKhorkov/libs/logging/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create chat window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockAuthWindow := mockwindows.NewMockWindow(ctrl)
			mockCreateChatWindow := mockwindows.NewMockWindow(ctrl)
			mockSearchUsersWindow := mockwindows.NewMockWindow(ctrl)
			mockNotificationWindow := mockwindows.NewMockWindow(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			w := New(
				app,
				mockAuthWindow,
				mockCreateChatWindow,
				mockSearchUsersWindow,
				mockNotificationWindow,
				mockLogger,
				mockUseCases,
			)

			assert.NotNil(t, w)
			assert.Equal(t, app, w.app)
			assert.Equal(t, mockUseCases, w.useCases)
			assert.Equal(t, mockAuthWindow, w.authWindow)
			assert.Equal(t, mockCreateChatWindow, w.createChatWindow)
			assert.Equal(t, mockSearchUsersWindow, w.searchUsersWindow)
			assert.Equal(t, mockNotificationWindow, w.notificationWindow)
			assert.Equal(t, mockLogger, w.logger)
			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_Build(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		setupMocks  func(*mockusecases.MockUseCases, *mocks.MockLogger)
		expectError bool
	}{
		{
			name: "successful build",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases, mockLogger *mocks.MockLogger) {
				mockUseCases.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{
						ID:        1,
						Username:  "testuser",
						Email:     "test@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				mockUseCases.EXPECT().
					GetUserChats(gomock.Any(), chatsLimit, chatsOffset).
					Return([]domains.Chat{}, nil)
			},
			expectError: false,
		},
		{
			name: "build with error getting current user",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases, mockLogger *mocks.MockLogger) {
				mockUseCases.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(nil, errors.New("user not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "ошибка получения пользователя", gomock.Any()).
					Times(1)
			},
			expectError: true,
		},
		{
			name: "build with error getting user chats",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases, mockLogger *mocks.MockLogger) {
				mockUseCases.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{
						ID:        1,
						Username:  "testuser",
						Email:     "test@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				mockUseCases.EXPECT().
					GetUserChats(gomock.Any(), chatsLimit, chatsOffset).
					Return(nil, errors.New("failed to load chats"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "ошибка загрузки чатов", gomock.Any()).
					Times(1)
			},
			expectError: false, // окно все равно строится
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockAuthWindow := mockwindows.NewMockWindow(ctrl)
			mockCreateChatWindow := mockwindows.NewMockWindow(ctrl)
			mockSearchUsersWindow := mockwindows.NewMockWindow(ctrl)
			mockNotificationWindow := mockwindows.NewMockWindow(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockLogger)
			}

			w := New(
				app,
				mockAuthWindow,
				mockCreateChatWindow,
				mockSearchUsersWindow,
				mockNotificationWindow,
				mockLogger,
				mockUseCases,
			)

			w.Build(nil)

			if !tt.expectError {
				assert.NotNil(t, w.window)
				assert.Equal(t, title, w.window.Title())
				assert.Equal(t, fyne.NewSize(width, height), w.window.Canvas().Size())
				assert.NotNil(t, w.window.Content())
			}
		})
	}
}

func TestWindow_Show(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		setupWindow func(*Window)
	}{
		{
			name: "show window after build",
			setupWindow: func(w *Window) {
				w.Build(nil)
			},
		},
		{
			name: "show without building",
			setupWindow: func(w *Window) {
				// Не вызываем Build
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockAuthWindow := mockwindows.NewMockWindow(ctrl)
			mockCreateChatWindow := mockwindows.NewMockWindow(ctrl)
			mockSearchUsersWindow := mockwindows.NewMockWindow(ctrl)
			mockNotificationWindow := mockwindows.NewMockWindow(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			// Настраиваем моки для успешного Build
			mockUseCases.EXPECT().
				GetCurrentUser(gomock.Any()).
				Return(&domains.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil).AnyTimes()

			mockUseCases.EXPECT().
				GetUserChats(gomock.Any(), chatsLimit, chatsOffset).
				Return([]domains.Chat{}, nil).AnyTimes()

			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

			w := New(
				app,
				mockAuthWindow,
				mockCreateChatWindow,
				mockSearchUsersWindow,
				mockNotificationWindow,
				mockLogger,
				mockUseCases,
			)

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			assert.NotPanics(t, func() {
				w.Show()
			})

			// Закрывавем окно, чтобы не было проблем с горутинами в других тестах
			w.Close()
		})
	}
}

func TestWindow_Close(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		setupWindow func(*Window)
	}{
		{
			name: "close existing window",
			setupWindow: func(w *Window) {
				w.Build(nil)
			},
		},
		{
			name: "close without building",
			setupWindow: func(w *Window) {
				// Не создаем окно
			},
		},
		{
			name: "close already closed window",
			setupWindow: func(w *Window) {
				w.Build(nil)
				w.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockAuthWindow := mockwindows.NewMockWindow(ctrl)
			mockCreateChatWindow := mockwindows.NewMockWindow(ctrl)
			mockSearchUsersWindow := mockwindows.NewMockWindow(ctrl)
			mockNotificationWindow := mockwindows.NewMockWindow(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			// Настраиваем моки для успешного Build
			mockUseCases.EXPECT().
				GetCurrentUser(gomock.Any()).
				Return(&domains.User{
					ID:        1,
					Username:  "testuser",
					Email:     "test@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil).AnyTimes()

			mockUseCases.EXPECT().
				GetUserChats(gomock.Any(), chatsLimit, chatsOffset).
				Return([]domains.Chat{}, nil).AnyTimes()

			w := New(
				app,
				mockAuthWindow,
				mockCreateChatWindow,
				mockSearchUsersWindow,
				mockNotificationWindow,
				mockLogger,
				mockUseCases,
			)

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			assert.NotPanics(t, func() {
				w.Close()
			})

			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant any
		expected any
	}{
		{
			name:     "title constant",
			constant: title,
			expected: "KFC Chat",
		},
		{
			name:     "width constant",
			constant: width,
			expected: 900,
		},
		{
			name:     "height constant",
			constant: height,
			expected: 700,
		},
		{
			name:     "chats limit",
			constant: chatsLimit,
			expected: 0,
		},
		{
			name:     "chats offset",
			constant: chatsOffset,
			expected: 0,
		},
		{
			name:     "chats list chat label text",
			constant: chatsListChatLabelText,
			expected: "chat",
		},
		{
			name:     "chats label text",
			constant: chatsLabelText,
			expected: "Чаты",
		},
		{
			name:     "current user sender label text",
			constant: currentUserSenderLabelText,
			expected: "Вы",
		},
		{
			name:     "messages header label text",
			constant: messagesHeaderLabelText,
			expected: "Сообщения",
		},
		{
			name:     "new chat button text",
			constant: newChatButtonText,
			expected: "Новый чат",
		},
		{
			name:     "search button text",
			constant: searchButtonText,
			expected: "Поиск",
		},
		{
			name:     "logout button text",
			constant: logoutButtonText,
			expected: "Выйти",
		},
		{
			name:     "load more messages button text",
			constant: loadMoreMessagesButtonText,
			expected: "Загрузить историю",
		},
		{
			name:     "close chat button text",
			constant: closeChatButtonText,
			expected: "Закрыть чат",
		},
		{
			name:     "send message button text",
			constant: sendMessageButtonText,
			expected: "Отправить",
		},
		{
			name:     "message entry text",
			constant: messageEntryText,
			expected: "Введите сообщение...",
		},
		{
			name:     "messages limit",
			constant: messagesLimit,
			expected: 10,
		},
		{
			name:     "refresh tokens interval",
			constant: refreshTokensInterval,
			expected: 1 * time.Minute,
		},
		{
			name:     "update chats interval",
			constant: updateChatsInterval,
			expected: 5 * time.Second,
		},
		{
			name:     "panels split offset",
			constant: panelsSplitOffset,
			expected: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// Интеграционный тест полного цикла.
func TestWindow_Integration(t *testing.T) {
	t.Run("full window lifecycle", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)
		mockAuthWindow := mockwindows.NewMockWindow(ctrl)
		mockCreateChatWindow := mockwindows.NewMockWindow(ctrl)
		mockSearchUsersWindow := mockwindows.NewMockWindow(ctrl)
		mockNotificationWindow := mockwindows.NewMockWindow(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)

		now := time.Now()

		// Настраиваем моки
		mockUseCases.EXPECT().
			GetCurrentUser(gomock.Any()).
			Return(&domains.User{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

		mockUseCases.EXPECT().
			GetUserChats(gomock.Any(), chatsLimit, chatsOffset).
			Return([]domains.Chat{
				{
					ID:        1,
					Type:      domains.ChatTypePrivate,
					IsRead:    true,
					CreatedAt: now,
					UpdatedAt: now,
					Members: []domains.User{
						{ID: 1, Username: "testuser"},
						{ID: 2, Username: "otheruser"},
					},
				},
			}, nil)

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(3)

		w := New(
			app,
			mockAuthWindow,
			mockCreateChatWindow,
			mockSearchUsersWindow,
			mockNotificationWindow,
			mockLogger,
			mockUseCases,
		)

		// Строим окно
		w.Build(nil)
		assert.NotNil(t, w.window)

		// Показываем
		w.Show()

		// Закрываем
		w.Close()
		assert.Nil(t, w.window)

		// Повторное закрытие не должно вызывать панику
		assert.NotPanics(t, func() {
			w.Close()
		})
	})
}
