package entries

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestMultilineEntry_TypedKey(t *testing.T) {
	t.Parallel() // Разрешаем параллельный запуск этого теста

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
			keyToPress:   "",
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
			inputText:    "Return key test",
			keyToPress:   fyne.KeyReturn,
			shouldSubmit: true,
			expectedText: "Return key test",
		},
	}

	for _, tt := range tests {
		// Запуск подтеста в параллели
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Инициализируем изолированное окружение Fyne для этого подтеста
			test.NewApp()

			entry := NewMultiLineEntry()

			var submittedText string

			wasSubmitted := false

			entry.OnSubmitted = func(s string) {
				wasSubmitted = true
				submittedText = s
			}

			// Вводим текст
			if tt.inputText != "" {
				test.Type(entry, tt.inputText)
			}

			// Проверяем нажатие клавиши
			if tt.keyToPress != "" {
				entry.FocusGained()
				entry.TypedKey(&fyne.KeyEvent{Name: tt.keyToPress})
				entry.FocusLost()
			}

			// Ассерты
			assert.Equal(t, tt.expectedText, entry.Text)
			assert.Equal(t, tt.shouldSubmit, wasSubmitted)

			if tt.shouldSubmit {
				assert.Equal(t, tt.expectedText, submittedText)
			}
		})
	}
}
