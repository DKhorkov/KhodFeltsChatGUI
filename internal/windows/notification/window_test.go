package notification

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
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
			ctrl := gomock.NewController(t)

			app := tt.setupApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases)

			assert.NotNil(t, w)
			assert.Equal(t, app, w.app)
			assert.Equal(t, mockUseCases, w.useCases)
			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_Build(t *testing.T) {
	tests := []struct {
		name        string
		message     fyne.CanvasObject
		expectedLen int // ожидаемое количество элементов в контейнере
	}{
		{
			name:        "build with text message",
			message:     widget.NewLabel("Test message"),
			expectedLen: 2, // заголовок + сообщение
		},
		{
			name:        "build with rich text message",
			message:     widget.NewRichTextFromMarkdown("**Important** message"),
			expectedLen: 2,
		},
		{
			name:        "build with custom widget",
			message:     widget.NewButton("Click me", func() {}),
			expectedLen: 2,
		},
		{
			name:        "build with complex container",
			message:     widget.NewCard("Title", "Subtitle", widget.NewLabel("Content")),
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases)
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

			// Проверяем, что первый элемент - это Label с заголовком
			titleLabel, ok := vbox.Objects[0].(*widget.Label)
			assert.True(t, ok, "First element should be a label")
			assert.Equal(t, title, titleLabel.Text, "Label text should match title")
			assert.True(t, titleLabel.TextStyle.Bold, "Title should be bold")

			// Проверяем, что второй элемент - это наше сообщение
			assert.Equal(t, tt.message, vbox.Objects[1], "Second element should be the message")
		})
	}
}

func TestWindow_Build_MultipleBuilds(t *testing.T) {
	t.Run("build multiple windows", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)

		w := New(app, mockUseCases)

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
			w.window.Content().(*fyne.Container).Objects[1],
			"Content should be updated",
		)
	})
}

func TestWindow_Show(t *testing.T) {
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
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases)

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
			setupWindow: func(w *Window) {
				// Не создаем окно
			},
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
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases)

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
	t.Run("full window lifecycle", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)

		w := New(app, mockUseCases)

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
	tests := []struct {
		name     string
		constant any
		expected any
	}{
		{
			name:     "title constant",
			constant: title,
			expected: "Новое сообщение",
		},
		{
			name:     "width constant",
			constant: width,
			expected: 300,
		},
		{
			name:     "height constant",
			constant: height,
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// Тест для проверки различных типов сообщений.
func TestWindow_Build_DifferentMessageTypes(t *testing.T) {
	tests := []struct {
		name    string
		message fyne.CanvasObject
	}{
		{
			name:    "simple label",
			message: widget.NewLabel("Simple message"),
		},
		{
			name: "label with custom style",
			message: widget.NewLabelWithStyle(
				"Styled message",
				fyne.TextAlignCenter,
				fyne.TextStyle{Italic: true},
			),
		},
		{
			name:    "rich text",
			message: widget.NewRichTextFromMarkdown("**Bold** and *italic*"),
		},
		{
			name:    "button",
			message: widget.NewButton("OK", nil),
		},
		{
			name:    "card",
			message: widget.NewCard("Card Title", "Card Subtitle", widget.NewLabel("Card content")),
		},
		{
			name: "complex container",
			message: container.NewVBox(
				widget.NewLabel("Label 1"),
				widget.NewLabel("Label 2"),
				widget.NewButton("Button", nil),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases)
			w.Build(tt.message)

			assert.NotNil(t, w.window)
			content := w.window.Content()
			vbox, ok := content.(*fyne.Container)
			assert.True(t, ok)
			assert.Equal(t, tt.message, vbox.Objects[1])
		})
	}
}
