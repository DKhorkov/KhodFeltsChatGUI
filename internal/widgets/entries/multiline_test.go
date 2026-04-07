package entries

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiLineEntry_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "инициализация multiline entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			test.NewApp() // Создаем контекст приложения для каждого теста

			entry := NewMultiLineEntry()

			assert.NotNil(t, entry)
			assert.True(t, entry.MultiLine)
		})
	}
}

func TestMultilineEntry_TypedKey(t *testing.T) {
	// Мы не используем t.Parallel() здесь, так как fyne.CurrentApp()
	// может вести себя нестабильно при параллельном создании приложений в тестах

	tests := []struct {
		name         string
		inputText    string
		keyToPress   fyne.KeyName
		shouldSubmit bool
		expectedText string
	}{
		{
			name:         "обычный ввод текста без нажатия Enter",
			inputText:    "Hello World",
			keyToPress:   "", // Ничего не нажимаем
			shouldSubmit: false,
			expectedText: "Hello World",
		},
		{
			name:         "отправка сообщения через KeyEnter",
			inputText:    "Message to send",
			keyToPress:   fyne.KeyEnter,
			shouldSubmit: true,
			expectedText: "Message to send",
		},
		{
			name:         "отправка сообщения через KeyReturn",
			inputText:    "Another message",
			keyToPress:   fyne.KeyReturn,
			shouldSubmit: true,
			expectedText: "Another message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.NewApp()
			entry := NewMultiLineEntry()

			var submittedText string
			wasSubmitted := false

			entry.OnSubmitted = func(s string) {
				wasSubmitted = true
				submittedText = s
			}

			// 1. Имитируем ввод текста
			if tt.inputText != "" {
				test.Type(entry, tt.inputText)
			}

			// 2. Имитируем нажатие специальной клавиши
			if tt.keyToPress != "" {
				entry.FocusGained()
				entry.TypedKey(&fyne.KeyEvent{Name: tt.keyToPress})
				entry.FocusLost()
			}

			// 3. Проверки
			assert.Equal(t, tt.expectedText, entry.Text)
			assert.Equal(t, tt.shouldSubmit, wasSubmitted, "Статус отправки не совпадает")

			if tt.shouldSubmit {
				assert.Equal(t, tt.expectedText, submittedText)
			}
		})
	}
}
