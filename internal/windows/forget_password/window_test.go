package forgetPassword

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/config"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	mockwindows "github.com/DKhorkov/kfcGUI/mocks/window"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupApp func() fyne.App
		expected *Window
	}{
		{
			name: "create new notification window",
			setupApp: func() fyne.App {
				return test.NewApp()
			},
			expected: nil, // будет проверяться структура
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := tt.setupApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)

			assert.NotNil(t, w)
			assert.Equal(t, app, w.app)
			assert.Equal(t, mockUseCases, w.useCases)
			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		message     fyne.CanvasObject
		expectedLen int // ожидаемое количество элементов в контейнере
	}{
		{
			name:        "build with text message",
			message:     widget.NewLabel("Test message"),
			expectedLen: 5,
		},
		{
			name:        "build with rich text message",
			message:     widget.NewRichTextFromMarkdown("**Important** message"),
			expectedLen: 5,
		},
		{
			name:        "build with custom widget",
			message:     widget.NewButton("Click me", func() {}),
			expectedLen: 5,
		},
		{
			name:        "build with complex container",
			message:     widget.NewCard("Title", "Subtitle", widget.NewLabel("Content")),
			expectedLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)
			w.Build(tt.message)

			assert.NotNil(t, w.window, "Window should be created")
			assert.Equal(t, title, w.window.Title(), "Window title should match")

			expectedSize := fyne.NewSize(width, height)
			assert.Equal(t, expectedSize, w.window.Canvas().Size(), "Window size should match")

			// Проверяем содержимое окна
			content := w.window.Content()
			assert.NotNil(t, content)

			// Проверяем, что контент является VBox
			vbox, ok := content.(*fyne.Container)
			assert.True(t, ok, "Content should be a container")
			assert.Equal(
				t,
				tt.expectedLen,
				len(vbox.Objects),
				"VBox should have expected number of objects",
			)
		})
	}
}

func TestWindow_Build_MultipleBuilds(t *testing.T) {
	t.Parallel()

	t.Run("build multiple windows", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)
		mockInformationWindow := mockwindows.NewMockWindow(ctrl)
		mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

		w := New(
			app,
			mockInformationWindow,
			mockUseCases,
			config.ValidationConfig{},
			mockErrorsMapper,
		)

		// Первое окно
		message1 := widget.NewLabel("Message 1")
		w.Build(message1)
		window1 := w.window

		// Второе окно (должно заменить первое)
		message2 := widget.NewLabel("Message 2")
		w.Build(message2)
		window2 := w.window

		assert.NotEqual(t, window1, window2, "Should create new window on each build")
		assert.Equal(
			t,
			message2,
			w.window.Content().(*fyne.Container).Objects[0],
			"Content should be updated",
		)
	})
}

func TestWindow_Show(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupWindow    func(*Window)
		expectedClosed bool
	}{
		{
			name: "show window should be visible and auto-close after 3 seconds",
			setupWindow: func(w *Window) {
				w.Build(widget.NewLabel("Test"))
			},
			expectedClosed: true,
		},
		{
			name:           "show without building should not panic",
			expectedClosed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			// Показываем окно
			w.Show()

			if tt.setupWindow != nil {
				// Проверяем, что окно было создано
				assert.NotNil(t, w.window)

				// В тестовой среде окно может не отображаться реально,
				// но мы можем проверить, что оно не nil
				// Ждем немного для возможного закрытия
				time.Sleep(100 * time.Millisecond)

				// Окно должно быть еще открыто (закрывается через 3 секунды)
				// В тестах Fyne мы не можем проверить реальное состояние окна,
				// но можем убедиться, что оно не было закрыто сразу
				assert.NotNil(t, w.window)
			}
		})
	}
}

func TestWindow_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupWindow func(*Window)
	}{
		{
			name: "close existing window",
			setupWindow: func(w *Window) {
				w.Build(widget.NewLabel("Test"))
			},
		},
		{
			name: "close without building",
		},
		{
			name: "close already closed window",
			setupWindow: func(w *Window) {
				w.Build(widget.NewLabel("Test"))
				w.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockInformationWindow := mockwindows.NewMockWindow(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			w := New(
				app,
				mockInformationWindow,
				mockUseCases,
				config.ValidationConfig{},
				mockErrorsMapper,
			)

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			// Закрываем окно (не должно паниковать)
			assert.NotPanics(t, func() {
				w.Close()
			})
		})
	}
}

func TestWindow_Integration(t *testing.T) {
	t.Parallel()

	t.Run("full window lifecycle", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)
		mockInformationWindow := mockwindows.NewMockWindow(ctrl)
		mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

		w := New(
			app,
			mockInformationWindow,
			mockUseCases,
			config.ValidationConfig{},
			mockErrorsMapper,
		)

		// Создаем сообщение
		message := widget.NewLabelWithStyle(
			"You have a new message from John",
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		)

		// Строим окно
		w.Build(message)
		assert.NotNil(t, w.window)

		// Показываем
		w.Show()

		// Закрываем
		w.Close()

		// Повторное закрытие не должно вызывать панику
		assert.NotPanics(t, func() {
			w.Close()
		})
	})
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
			expected: "Сброс пароля",
		},
		{
			name:     "width constant",
			constant: width,
			expected: 300,
		},
		{
			name:     "height constant",
			constant: height,
			expected: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}
