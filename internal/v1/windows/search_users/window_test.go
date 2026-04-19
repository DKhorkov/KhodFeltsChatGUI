package searchUsers

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
	tests := []struct {
		name string
	}{
		{
			name: "create new search users window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
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
		name string
	}{
		{
			name: "build window with all components",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			w := New(app, mockUseCases)
			w.Build(nil)

			assert.NotNil(t, w.window)
			assert.Equal(t, title, w.window.Title())
			assert.Equal(t, fyne.NewSize(width, height), w.window.Canvas().Size())
			assert.NotNil(t, w.window.Content())
		})
	}
}

func TestWindow_SearchFunctionality(t *testing.T) {
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
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			w := New(app, mockUseCases)
			w.Build(nil)

			// Ищем searchEntry
			searchEntry := findSearchEntry(w.window)
			assert.NotNil(t, searchEntry)

			if tt.searchQuery != "" {
				searchEntry.SetText(tt.searchQuery)
				// Имитируем нажатие Enter
				searchEntry.OnSubmitted(tt.searchQuery)
			}

			// Даем время для выполнения горутины
			time.Sleep(100 * time.Millisecond)

			// Проверяем, что не было паники
			assert.NotPanics(t, func() {
				w.Close()
			})
		})
	}
}

func TestWindow_SearchButton(t *testing.T) {
	tests := []struct {
		name        string
		searchQuery string
		setupMocks  func(*mockusecases.MockUseCases)
	}{
		{
			name:        "search button click with query",
			searchQuery: "test_user",
			setupMocks: func(mockUseCases *mockusecases.MockUseCases) {
				mockUseCases.EXPECT().
					SearchUsers(gomock.Any(), "test_user", limit, offset).
					Return([]domains.User{
						{ID: 1, Username: "test_user", Email: "test@example.com"},
					}, nil)
			},
		},
		{
			name:        "search button click with empty query",
			searchQuery: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			w := New(app, mockUseCases)
			w.Build(nil)

			searchEntry := findSearchEntry(w.window)
			searchButton := findSearchButton(w.window)

			if tt.searchQuery != "" {
				searchEntry.SetText(tt.searchQuery)
			}

			if searchButton != nil {
				test.Tap(searchButton)
			}

			// Даем время для выполнения горутины
			time.Sleep(100 * time.Millisecond)
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

			w := New(app, mockUseCases)

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

			w := New(app, mockUseCases)

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
			expected: "Поиск пользователей",
		},
		{
			name:     "width constant",
			constant: width,
			expected: 400,
		},
		{
			name:     "height constant",
			constant: height,
			expected: 500,
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
			name:     "username label index",
			constant: usernameLabelIndex,
			expected: 1,
		},
		{
			name:     "email label index",
			constant: emailLabelIndex,
			expected: 2,
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

func TestWindow_UsersList(t *testing.T) {
	tests := []struct {
		name       string
		users      []domains.User
		setupMocks func(*mockusecases.MockUseCases)
	}{
		{
			name: "list with multiple users",
			users: []domains.User{
				{ID: 1, Username: "user1", Email: "user1@example.com"},
				{ID: 2, Username: "user2", Email: "user2@example.com"},
				{ID: 3, Username: "user3", Email: "user3@example.com"},
			},
			setupMocks: func(mockUseCases *mockusecases.MockUseCases) {
				mockUseCases.EXPECT().
					SearchUsers(gomock.Any(), "test", limit, offset).
					Return([]domains.User{
						{ID: 1, Username: "user1", Email: "user1@example.com"},
						{ID: 2, Username: "user2", Email: "user2@example.com"},
						{ID: 3, Username: "user3", Email: "user3@example.com"},
					}, nil)
			},
		},
		{
			name:  "empty list",
			users: []domains.User{},
			setupMocks: func(mockUseCases *mockusecases.MockUseCases) {
				mockUseCases.EXPECT().
					SearchUsers(gomock.Any(), "test", limit, offset).
					Return([]domains.User{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := test.NewApp()
			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			w := New(app, mockUseCases)
			w.Build(nil)

			// Имитируем поиск
			searchEntry := findSearchEntry(w.window)
			assert.NotNil(t, searchEntry)
			searchEntry.SetText("test")
			searchEntry.OnSubmitted("test")

			// Даем время для выполнения горутины
			time.Sleep(100 * time.Millisecond)

			// Проверяем, что окно существует
			assert.NotNil(t, w.window)
		})
	}
}

// Вспомогательные функции для поиска виджетов в тестах.
func findSearchEntry(window fyne.Window) *widget.Entry {
	if window == nil || window.Content() == nil {
		return nil
	}

	var entry *widget.Entry

	// Рекурсивный обход дерева виджетов
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

func findSearchButton(window fyne.Window) *widget.Button {
	if window == nil || window.Content() == nil {
		return nil
	}

	var button *widget.Button

	var walk func(w fyne.CanvasObject)

	walk = func(w fyne.CanvasObject) {
		if b, ok := w.(*widget.Button); ok {
			if b.Text == searchButtonName {
				button = b

				return
			}
		}

		if container, ok := w.(*fyne.Container); ok {
			for _, child := range container.Objects {
				walk(child)

				if button != nil {
					return
				}
			}
		}
	}

	walk(window.Content())

	return button
}

func TestWindow_Integration(t *testing.T) {
	t.Run("full window lifecycle", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		app := test.NewApp()
		mockUseCases := mockusecases.NewMockUseCases(ctrl)

		// Настраиваем мок для успешного поиска
		mockUseCases.EXPECT().
			SearchUsers(gomock.Any(), "test_user", limit, offset).
			Return([]domains.User{
				{ID: 1, Username: "test_user", Email: "test@example.com"},
			}, nil)

		w := New(app, mockUseCases)

		// Строим окно
		w.Build(nil)
		assert.NotNil(t, w.window)

		// Показываем
		w.Show()

		// Имитируем поиск
		searchEntry := findSearchEntry(w.window)
		if searchEntry != nil {
			searchEntry.SetText("test_user")
			searchEntry.OnSubmitted("test_user")
		}

		// Даем время для выполнения горутины
		time.Sleep(100 * time.Millisecond)

		// Закрываем
		w.Close()
		assert.Nil(t, w.window)

		// Повторное закрытие не должно вызывать панику
		assert.NotPanics(t, func() {
			w.Close()
		})
	})
}
