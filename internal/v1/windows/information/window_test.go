package information

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected *Window
	}{
		{
			name:     "create window with title",
			title:    "Information",
			expected: nil,
		},
		{
			name:     "create window with empty title",
			title:    "",
			expected: nil,
		},
		{
			name:     "create window with long title",
			title:    "This is a very long title that might wrap",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := test.NewApp()

			w := New(app, tt.title)

			assert.NotNil(t, w)
			assert.Equal(t, app, w.app)
			assert.Equal(t, tt.title, w.title)
			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_Build(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		content     fyne.CanvasObject
		expectedLen int
	}{
		{
			name:        "build with label content",
			title:       "Info",
			content:     widget.NewLabel("This is information"),
			expectedLen: 0,
		},
		{
			name:        "build with button content",
			title:       "Confirm",
			content:     widget.NewButton("OK", nil),
			expectedLen: 0,
		},
		{
			name:  "build with complex container",
			title: "Details",
			content: container.NewVBox(
				widget.NewLabel("Label 1"),
				widget.NewLabel("Label 2"),
				widget.NewButton("Action", nil),
			),
			expectedLen: 0,
		},
		{
			name:        "build with card",
			title:       "Card Info",
			content:     widget.NewCard("Title", "Subtitle", widget.NewLabel("Content")),
			expectedLen: 0,
		},
		{
			name:        "build with entry widget",
			title:       "Input",
			content:     widget.NewEntry(),
			expectedLen: 0,
		},
		{
			name:        "build with empty title",
			title:       "",
			content:     widget.NewLabel("Message"),
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := test.NewApp()

			w := New(app, tt.title)
			w.Build(tt.content)

			assert.NotNil(t, w.window, "Window should be created")
			assert.Equal(t, tt.title, w.window.Title(), "Window title should match")

			expectedSize := fyne.NewSize(width, height)
			assert.Equal(t, expectedSize, w.window.Canvas().Size(), "Window size should match")

			content := w.window.Content()
			assert.NotNil(t, content)
			assert.Equal(t, tt.content, content, "Window content should match the provided content")
		})
	}
}

func TestWindow_Build_MultipleBuilds(t *testing.T) {
	t.Run("build multiple windows sequentially", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "First Window")

		// Первое окно
		content1 := widget.NewLabel("First content")
		w.Build(content1)
		window1 := w.window

		// Второе окно (должно заменить первое)
		content2 := widget.NewLabel("Second content")
		w.Build(content2)
		window2 := w.window

		assert.NotEqual(t, window1, window2, "Should create new window on each build")
		assert.Equal(t, content2, w.window.Content(), "Content should be updated")
	})
}

func TestWindow_Show(t *testing.T) {
	tests := []struct {
		name        string
		setupWindow func(*Window)
	}{
		{
			name: "show window after build",
			setupWindow: func(w *Window) {
				w.Build(widget.NewLabel("Test content"))
			},
		},
		{
			name: "show without building",
		},
		{
			name: "show after close",
			setupWindow: func(w *Window) {
				w.Build(widget.NewLabel("Test"))
				w.Close()
				// Пытаемся показать после закрытия
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := test.NewApp()

			w := New(app, "Test")

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			// Показываем окно (не должно паниковать)
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
		{
			name: "close multiple times",
			setupWindow: func(w *Window) {
				w.Build(widget.NewLabel("Test"))
				w.Close()
				w.Close()
				w.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := test.NewApp()

			w := New(app, "Test")

			if tt.setupWindow != nil {
				tt.setupWindow(w)
			}

			// Закрываем окно (не должно паниковать)
			assert.NotPanics(t, func() {
				w.Close()
			})

			// После закрытия window должен быть nil
			assert.Nil(t, w.window, "Window should be nil after close")
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
			t.Parallel()

			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestWindow_Integration(t *testing.T) {
	t.Run("full window lifecycle", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Integration Test")

		// Создаем контент
		content := widget.NewLabelWithStyle(
			"Test information message",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		// Строим окно
		w.Build(content)
		assert.NotNil(t, w.window)
		assert.Equal(t, "Integration Test", w.window.Title())

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

func TestWindow_Build_DifferentContentTypes(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content fyne.CanvasObject
	}{
		{
			name:    "label widget",
			title:   "Label Window",
			content: widget.NewLabel("Simple label"),
		},
		{
			name:    "entry widget",
			title:   "Entry Window",
			content: widget.NewEntry(),
		},
		{
			name:    "button widget",
			title:   "Button Window",
			content: widget.NewButton("Click me", nil),
		},
		{
			name:    "check widget",
			title:   "Check Window",
			content: widget.NewCheck("Option", nil),
		},
		{
			name:    "progress bar",
			title:   "Progress Window",
			content: widget.NewProgressBar(),
		},
		{
			name:  "form widget",
			title: "Form Window",
			content: widget.NewForm(
				widget.NewFormItem("Name", widget.NewEntry()),
				widget.NewFormItem("Email", widget.NewEntry()),
			),
		},
		{
			name:  "accordion",
			title: "Accordion Window",
			content: widget.NewAccordion(
				widget.NewAccordionItem("Item 1", widget.NewLabel("Content 1")),
				widget.NewAccordionItem("Item 2", widget.NewLabel("Content 2")),
			),
		},
		{
			name:  "list widget",
			title: "List Window",
			content: widget.NewList(
				func() int { return 3 },
				func() fyne.CanvasObject { return widget.NewLabel("") },
				func(i int, o fyne.CanvasObject) { o.(*widget.Label).SetText("Item " + string(rune('1'+i))) },
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := test.NewApp()

			w := New(app, tt.title)
			w.Build(tt.content)

			assert.NotNil(t, w.window)
			assert.Equal(t, tt.title, w.window.Title())
			assert.Equal(t, tt.content, w.window.Content())
			assert.Equal(t, fyne.NewSize(width, height), w.window.Canvas().Size())
		})
	}
}

func TestWindow_ShowHideSequence(t *testing.T) {
	t.Run("show, close, show again", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Test")

		// Создаем и показываем окно
		firstContent := widget.NewLabel("Content")
		w.Build(firstContent)
		w.Show()
		assert.NotNil(t, w.window)

		// Проверяем, что содержимое соответствует
		assert.Equal(t, firstContent.Text, w.window.Content().(*widget.Label).Text)

		// Закрываем окно
		w.Close()
		assert.Nil(t, w.window)

		// Показываем после закрытия - не должно паниковать
		assert.NotPanics(t, func() {
			w.Show()
		})

		// Строим новое окно
		secondContent := widget.NewLabel("New Content")
		w.Build(secondContent)
		assert.NotNil(t, w.window)

		// Сравниваем только текст, а не всю структуру
		actualLabel, ok := w.window.Content().(*widget.Label)
		assert.True(t, ok)
		assert.Equal(t, secondContent.Text, actualLabel.Text)

		// Показываем новое окно
		assert.NotPanics(t, func() {
			w.Show()
		})
	})
}

func TestWindow_MultipleInstances(t *testing.T) {
	t.Run("create multiple independent windows", func(t *testing.T) {
		app := test.NewApp()

		// Создаем первое окно
		w1 := New(app, "Window 1")
		content1 := widget.NewLabel("Content 1")
		w1.Build(content1)

		// Создаем второе окно
		w2 := New(app, "Window 2")
		content2 := widget.NewLabel("Content 2")
		w2.Build(content2)

		// Проверяем, что окна независимы
		assert.NotEqual(t, w1.window, w2.window)
		assert.Equal(t, "Window 1", w1.window.Title())
		assert.Equal(t, "Window 2", w2.window.Title())

		// Проверяем содержимое первого окна
		label1, ok1 := w1.window.Content().(*widget.Label)
		assert.True(t, ok1)
		assert.Equal(t, "Content 1", label1.Text)

		// Проверяем содержимое второго окна
		label2, ok2 := w2.window.Content().(*widget.Label)
		assert.True(t, ok2)
		assert.Equal(t, "Content 2", label2.Text)

		// Закрываем первое окно
		w1.Close()
		assert.Nil(t, w1.window)
		assert.NotNil(t, w2.window)

		// Закрываем второе окно
		w2.Close()
		assert.Nil(t, w2.window)
	})
}

// Дополнительный тест для проверки разных типов содержимого после перестроения.
func TestWindow_RebuildWithDifferentContent(t *testing.T) {
	t.Run("rebuild window with different content types", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Test Window")

		// Создаем окно с Label
		labelContent := widget.NewLabel("Label text")
		w.Build(labelContent)
		assert.NotNil(t, w.window)

		// Проверяем, что содержимое - Label с правильным текстом
		label, ok := w.window.Content().(*widget.Label)
		assert.True(t, ok)
		assert.Equal(t, "Label text", label.Text)

		// Закрываем
		w.Close()
		assert.Nil(t, w.window)

		// Создаем новое окно с Button
		buttonContent := widget.NewButton("Button text", nil)
		w.Build(buttonContent)
		assert.NotNil(t, w.window)

		// Проверяем, что содержимое - Button с правильным текстом
		button, ok := w.window.Content().(*widget.Button)
		assert.True(t, ok)
		assert.Equal(t, "Button text", button.Text)

		// Закрываем
		w.Close()
		assert.Nil(t, w.window)

		// Создаем новое окно с Entry
		entryContent := widget.NewEntry()
		entryContent.SetText("Entry text")
		w.Build(entryContent)
		assert.NotNil(t, w.window)

		// Проверяем, что содержимое - Entry с правильным текстом
		entry, ok := w.window.Content().(*widget.Entry)
		assert.True(t, ok)
		assert.Equal(t, "Entry text", entry.Text)
	})
}

// Тест для проверки, что Close безопасно вызывается несколько раз.
func TestWindow_SafeClose(t *testing.T) {
	t.Run("close multiple times safely", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Test")

		// Закрываем без построения
		assert.NotPanics(t, func() {
			w.Close()
		})
		assert.Nil(t, w.window)

		// Строим и закрываем
		w.Build(widget.NewLabel("Content"))
		assert.NotNil(t, w.window)

		assert.NotPanics(t, func() {
			w.Close()
		})
		assert.Nil(t, w.window)

		// Закрываем снова
		assert.NotPanics(t, func() {
			w.Close()
		})
		assert.Nil(t, w.window)
	})
}

// Тест для проверки Show без Build.
func TestWindow_ShowWithoutBuild(t *testing.T) {
	t.Run("show without building should not panic", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Test")

		assert.NotPanics(t, func() {
			w.Show()
		})
		assert.Nil(t, w.window)

		// После Build должно работать нормально
		w.Build(widget.NewLabel("Content"))
		assert.NotPanics(t, func() {
			w.Show()
		})
		assert.NotNil(t, w.window)
	})
}

// Тест для проверки размера окна.
func TestWindow_WindowSize(t *testing.T) {
	t.Run("window size should match constants", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Test")
		w.Build(widget.NewLabel("Content"))

		expectedSize := fyne.NewSize(width, height)
		assert.Equal(t, expectedSize, w.window.Canvas().Size())
	})
}

// Тест для проверки заголовка окна.
func TestWindow_Title(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		expectedTitle string
	}{
		{
			name:          "normal title",
			title:         "Information Window",
			expectedTitle: "Information Window",
		},
		{
			name:          "empty title",
			title:         "",
			expectedTitle: "",
		},
		{
			name:          "unicode title",
			title:         "Информационное окно",
			expectedTitle: "Информационное окно",
		},
		{
			name:          "title with emoji",
			title:         "Info ℹ️",
			expectedTitle: "Info ℹ️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := test.NewApp()

			w := New(app, tt.title)
			w.Build(widget.NewLabel("Content"))

			assert.Equal(t, tt.expectedTitle, w.window.Title())
		})
	}
}

// Тест для проверки, что Build пересоздает окно с новым содержимым.
func TestWindow_BuildReplacesWindow(t *testing.T) {
	t.Run("build should create new window", func(t *testing.T) {
		app := test.NewApp()

		w := New(app, "Test")

		// Первый Build
		w.Build(widget.NewLabel("First"))
		firstWindow := w.window
		assert.NotNil(t, firstWindow)

		// Второй Build
		w.Build(widget.NewLabel("Second"))
		secondWindow := w.window
		assert.NotNil(t, secondWindow)

		// Должны быть разные окна
		assert.NotEqual(t, firstWindow, secondWindow)

		// Содержимое должно быть новым
		label, ok := secondWindow.Content().(*widget.Label)
		assert.True(t, ok)
		assert.Equal(t, "Second", label.Text)
		// Первое окно больше не должно быть доступно
		// (оно закрылось автоматически при создании нового)
	})
}
