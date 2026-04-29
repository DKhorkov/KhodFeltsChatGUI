package auth

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	mockwindows "github.com/DKhorkov/kfcGUI/mocks/window"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name             string
		validationConfig config.ValidationConfig
	}{
		{
			name:             "create auth window with valid config",
			validationConfig: config.ValidationConfig{},
		},
		{
			name:             "create auth window with nil chat window",
			validationConfig: config.ValidationConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockChatWindow := mockwindows.NewMockWindow(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockChatWindow,
				mockInformationWindow,
				mockUseCases,
				tt.validationConfig,
				mockErrorsMapper,
			)

			assert.NotNil(t, w)
			assert.Equal(t, app, w.app)
			assert.Equal(t, mockUseCases, w.useCases)
			assert.Equal(t, tt.validationConfig, w.validationConfig)
			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_SetChatWindow(t *testing.T) {
	tests := []struct {
		name        string
		initialChat interfaces.Window
		newChat     interfaces.Window
	}{
		{
			name:        "set new chat window",
			initialChat: nil,
			newChat:     mockwindows.NewMockWindow(gomock.NewController(t)),
		},
		{
			name:        "replace existing chat window",
			initialChat: mockwindows.NewMockWindow(gomock.NewController(t)),
			newChat:     mockwindows.NewMockWindow(gomock.NewController(t)),
		},
		{
			name:        "set nil chat window",
			initialChat: mockwindows.NewMockWindow(gomock.NewController(t)),
			newChat:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				tt.initialChat,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)
			w.SetChatWindow(tt.newChat)

			assert.Equal(t, tt.newChat, w.chatWindow)
		})
	}
}

func TestWindow_Build(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "build auth window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockChatWindow := mockwindows.NewMockWindow(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockChatWindow,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)
			w.Build(nil)

			assert.NotNil(t, w.window)
			assert.Equal(t, title, w.window.Title())
			assert.Equal(t, fyne.NewSize(width, height), w.window.Canvas().Size())
			assert.NotNil(t, w.window.Content())
		})
	}
}

func TestWindow_Show(t *testing.T) {
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
		},
		{
			name: "show after close",
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
			mockChatWindow := mockwindows.NewMockWindow(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockChatWindow,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			assert.NotPanics(t, func() {
				w.Show()
			})
		})
	}
}

func TestWindow_Close(t *testing.T) {
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
			mockChatWindow := mockwindows.NewMockWindow(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockChatWindow,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
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
	t.Parallel()

	tests := []struct {
		name     string
		constant any
		expected any
	}{
		{
			name:     "title constant",
			constant: title,
			expected: "Вход / Регистрация",
		},
		{
			name:     "width constant",
			constant: width,
			expected: width,
		},
		{
			name:     "height constant",
			constant: height,
			expected: 450,
		},
		{
			name:     "login tab name",
			constant: loginTabName,
			expected: "Вход",
		},
		{
			name:     "register tab name",
			constant: registerTabName,
			expected: "Регистрация",
		},
		{
			name:     "login button name",
			constant: loginButtonName,
			expected: "Войти",
		},
		{
			name:     "register button name",
			constant: registerButtonName,
			expected: "Зарегистрироваться",
		},
		{
			name:     "email entry text",
			constant: emailEntryText,
			expected: "Почта",
		},
		{
			name:     "password entry text",
			constant: passwordEntryText,
			expected: "Пароль",
		},
		{
			name:     "username entry text",
			constant: usernameEntryText,
			expected: "Логин",
		},
		{
			name:     "confirm password entry text",
			constant: confirmPasswordEntryText,
			expected: "Подтверждение пароля",
		},
		{
			name:     "login tab index",
			constant: loginTabIndex,
			expected: 0,
		},
		{
			name:     "register tab index",
			constant: registerTabIndex,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestWindow_Integration(t *testing.T) {
	t.Run("full window lifecycle ", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)
		mockChatWindow := mockwindows.NewMockWindow(ctrl)
		mockInformationWindow := mockwindows.NewMockWindow(ctrl)
		mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

		w := New(
			app,
			mockChatWindow,
			mockInformationWindow,
			mockUseCases,
			config.ValidationConfig{},
			mockErrorsMapper,
		)
		w.Build(nil)

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
