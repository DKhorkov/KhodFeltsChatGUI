package createChat

import (
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		refreshChatsFunc func(domains.Chat)
	}{
		{
			name:             "create window with refresh function",
			refreshChatsFunc: func(_ domains.Chat) {},
		},
		{
			name:             "create window with nil refresh function",
			refreshChatsFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases, tt.refreshChatsFunc)

			assert.NotNil(t, w)
			assert.Equal(t, app, w.app)
			assert.Equal(t, mockUseCases, w.useCases)
			// Проверяем, что функция установлена (не сравниваем функции напрямую)
			if tt.refreshChatsFunc != nil {
				assert.NotNil(t, w.refreshChatsFunc)
			} else {
				assert.Nil(t, w.refreshChatsFunc)
			}

			assert.Nil(t, w.window)
		})
	}
}

func TestWindow_SetRefreshChatsFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initialFunc func(domains.Chat)
		newFunc     func(domains.Chat)
	}{
		{
			name:        "set new refresh function",
			initialFunc: func(_ domains.Chat) {},
			newFunc:     func(_ domains.Chat) {},
		},
		{
			name:        "set nil refresh function",
			initialFunc: func(_ domains.Chat) {},
			newFunc:     nil,
		},
		{
			name:        "replace nil with function",
			initialFunc: nil,
			newFunc:     func(_ domains.Chat) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases, tt.initialFunc)
			w.SetRefreshChatsFunc(tt.newFunc)

			// Проверяем, что функция установлена (не сравниваем функции напрямую)
			if tt.newFunc != nil {
				assert.NotNil(t, w.refreshChatsFunc)
			} else {
				assert.Nil(t, w.refreshChatsFunc)
			}
		})
	}
}

func TestWindow_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "build window with all components",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases, nil)
			w.Build(nil)

			assert.NotNil(t, w.window)
			assert.Equal(t, title, w.window.Title())
			assert.Equal(t, fyne.NewSize(width, height), w.window.Canvas().Size())
			assert.NotNil(t, w.window.Content())
		})
	}
}

func TestWindow_SearchFunctionality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		searchQuery   string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedError bool
	}{
		{
			name:        "successful search with results",
			searchQuery: "john",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases) {
				mockUseCases.EXPECT().
					SearchUsers(gomock.Any(), "john", limit, offset).
					Return([]domains.User{
						{ID: 1, Username: "john_doe", Email: "john@example.com"},
						{ID: 2, Username: "john_smith", Email: "john.smith@example.com"},
					}, nil)
			},
			expectedError: false,
		},
		{
			name:        "search with empty results",
			searchQuery: "nonexistent",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases) {
				mockUseCases.EXPECT().
					SearchUsers(gomock.Any(), "nonexistent", limit, offset).
					Return([]domains.User{}, nil)
			},
			expectedError: false,
		},
		{
			name:          "search with empty query",
			searchQuery:   "",
			expectedError: false,
		},
		{
			name:        "search with error",
			searchQuery: "error",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases) {
				mockUseCases.EXPECT().
					SearchUsers(gomock.Any(), "error", limit, offset).
					Return(nil, errors.New("network error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			w := New(app, mockUseCases, nil)
			w.Build(nil)

			searchEntry := findSearchEntry(w.window)
			assert.NotNil(t, searchEntry)

			if tt.searchQuery != "" {
				searchEntry.SetText(tt.searchQuery)
				searchEntry.OnSubmitted(tt.searchQuery)
			}

			time.Sleep(100 * time.Millisecond)

			assert.NotPanics(t, func() {
				w.Close()
			})
		})
	}
}

func TestWindow_ChatTypeSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		selectedType string
		expectedType domains.ChatType
	}{
		{
			name:         "select private chat",
			selectedType: string(domains.ChatTypePrivate),
			expectedType: domains.ChatTypePrivate,
		},
		{
			name:         "select group chat",
			selectedType: string(domains.ChatTypeGroup),
			expectedType: domains.ChatTypeGroup,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases, nil)
			w.Build(nil)

			typeSelect := findTypeSelect(w.window)
			if typeSelect != nil {
				typeSelect.SetSelected(tt.selectedType)
				assert.Equal(t, tt.selectedType, typeSelect.Selected)
			}
		})
	}
}

func TestWindow_Show(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases, nil)

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
	t.Parallel()

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
			t.Parallel()

			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases, nil)

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
			expected: "Создать чат",
		},
		{
			name:     "width constant",
			constant: width,
			expected: 450,
		},
		{
			name:     "height constant",
			constant: height,
			expected: 400,
		},
		{
			name:     "search entry text",
			constant: searchEntryText,
			expected: "Введите имя пользователя...",
		},
		{
			name:     "search button name",
			constant: searchButtonName,
			expected: "Найти",
		},
		{
			name:     "create chat button name",
			constant: createChatButtonName,
			expected: "Создать",
		},
		{
			name:     "chat type label text",
			constant: chatTypeLabelText,
			expected: "Тип чата:",
		},
		{
			name:     "username label text",
			constant: usernameLabelText,
			expected: "Имя пользователя",
		},
		{
			name:     "email label text",
			constant: emailLabelText,
			expected: "email",
		},
		{
			name:     "checked label index",
			constant: checkedLabelIndex,
			expected: 1,
		},
		{
			name:     "username label index",
			constant: usernameLabelIndex,
			expected: 2,
		},
		{
			name:     "email label index",
			constant: emailLabelIndex,
			expected: 3,
		},
		{
			name:     "limit",
			constant: limit,
			expected: 0,
		},
		{
			name:     "offset",
			constant: offset,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestWindow_ErrorNoChatMembers(t *testing.T) {
	t.Parallel()

	t.Run("error message should be correct", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "укажите хотя бы одного участника", errNoChatMembersProvided.Error())
	})
}

// Вспомогательные функции для поиска виджетов.
func findSearchEntry(window fyne.Window) *widget.Entry {
	if window == nil || window.Content() == nil {
		return nil
	}

	var entry *widget.Entry

	var walk func(w fyne.CanvasObject)

	walk = func(w fyne.CanvasObject) {
		if e, ok := w.(*widget.Entry); ok {
			if e.PlaceHolder == searchEntryText {
				entry = e

				return
			}
		}

		if container, ok := w.(*fyne.Container); ok {
			for _, child := range container.Objects {
				walk(child)

				if entry != nil {
					return
				}
			}
		}
	}

	walk(window.Content())

	return entry
}

func findTypeSelect(window fyne.Window) *widget.Select {
	if window == nil || window.Content() == nil {
		return nil
	}

	var selectWidget *widget.Select

	var walk func(w fyne.CanvasObject)

	walk = func(w fyne.CanvasObject) {
		if s, ok := w.(*widget.Select); ok {
			selectWidget = s

			return
		}

		if container, ok := w.(*fyne.Container); ok {
			for _, child := range container.Objects {
				walk(child)

				if selectWidget != nil {
					return
				}
			}
		}
	}

	walk(window.Content())

	return selectWidget
}

func TestWindow_Integration(t *testing.T) {
	t.Parallel()

	t.Run("full window lifecycle ", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)

		refreshFunc := func(chat domains.Chat) {
			assert.Equal(t, uint64(1), chat.ID)
		}

		w := New(app, mockUseCases, refreshFunc)
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
