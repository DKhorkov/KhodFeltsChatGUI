package windows_test

import (
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/windows"
	"github.com/stretchr/testify/assert"
)

func TestFoundUsersText(t *testing.T) {
	tests := []struct {
		name       string
		usersCount int
		expected   string
	}{
		// Особые случаи (11-14)
		{
			name:       "11 users - exception",
			usersCount: 11,
			expected:   "Найдено 11 пользователей:",
		},
		{
			name:       "12 users - exception",
			usersCount: 12,
			expected:   "Найдено 12 пользователей:",
		},
		{
			name:       "13 users - exception",
			usersCount: 13,
			expected:   "Найдено 13 пользователей:",
		},
		{
			name:       "14 users - exception",
			usersCount: 14,
			expected:   "Найдено 14 пользователей:",
		},
		{
			name:       "111 users - exception (ends with 11)",
			usersCount: 111,
			expected:   "Найдено 111 пользователей:",
		},
		{
			name:       "112 users - exception (ends with 12)",
			usersCount: 112,
			expected:   "Найдено 112 пользователей:",
		},

		// Случаи с окончанием на 1 (кроме 11)
		{
			name:       "1 user",
			usersCount: 1,
			expected:   "Найден 1 пользователь:",
		},
		{
			name:       "21 users",
			usersCount: 21,
			expected:   "Найден 21 пользователь:",
		},
		{
			name:       "31 users",
			usersCount: 31,
			expected:   "Найден 31 пользователь:",
		},
		{
			name:       "101 users",
			usersCount: 101,
			expected:   "Найден 101 пользователь:",
		},

		// Случаи с окончанием на 2, 3, 4 (кроме 12, 13, 14)
		{
			name:       "2 users",
			usersCount: 2,
			expected:   "Найдено 2 пользователя:",
		},
		{
			name:       "3 users",
			usersCount: 3,
			expected:   "Найдено 3 пользователя:",
		},
		{
			name:       "4 users",
			usersCount: 4,
			expected:   "Найдено 4 пользователя:",
		},
		{
			name:       "22 users",
			usersCount: 22,
			expected:   "Найдено 22 пользователя:",
		},
		{
			name:       "23 users",
			usersCount: 23,
			expected:   "Найдено 23 пользователя:",
		},
		{
			name:       "24 users",
			usersCount: 24,
			expected:   "Найдено 24 пользователя:",
		},
		{
			name:       "102 users",
			usersCount: 102,
			expected:   "Найдено 102 пользователя:",
		},
		{
			name:       "103 users",
			usersCount: 103,
			expected:   "Найдено 103 пользователя:",
		},
		{
			name:       "104 users",
			usersCount: 104,
			expected:   "Найдено 104 пользователя:",
		},

		// Случаи с окончанием на 0, 5-9 (исключая особые случаи)
		{
			name:       "5 users",
			usersCount: 5,
			expected:   "Найдено 5 пользователей:",
		},
		{
			name:       "6 users",
			usersCount: 6,
			expected:   "Найдено 6 пользователей:",
		},
		{
			name:       "7 users",
			usersCount: 7,
			expected:   "Найдено 7 пользователей:",
		},
		{
			name:       "8 users",
			usersCount: 8,
			expected:   "Найдено 8 пользователей:",
		},
		{
			name:       "9 users",
			usersCount: 9,
			expected:   "Найдено 9 пользователей:",
		},
		{
			name:       "10 users",
			usersCount: 10,
			expected:   "Найдено 10 пользователей:",
		},
		{
			name:       "15 users",
			usersCount: 15,
			expected:   "Найдено 15 пользователей:",
		},
		{
			name:       "20 users",
			usersCount: 20,
			expected:   "Найдено 20 пользователей:",
		},
		{
			name:       "25 users",
			usersCount: 25,
			expected:   "Найдено 25 пользователей:",
		},
		{
			name:       "30 users",
			usersCount: 30,
			expected:   "Найдено 30 пользователей:",
		},
		{
			name:       "100 users",
			usersCount: 100,
			expected:   "Найдено 100 пользователей:",
		},
		{
			name:       "0 users",
			usersCount: 0,
			expected:   "Найдено 0 пользователей:",
		},

		// Крайние случаи
		{
			name:       "large number not ending with exception",
			usersCount: 999,
			expected:   "Найдено 999 пользователей:",
		},
		{
			name:       "large number ending with 1 but not exception",
			usersCount: 1001,
			expected:   "Найден 1001 пользователь:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := windows.FoundUsersText(tt.usersCount)
			assert.Equal(t, tt.expected, result)
		})
	}
}
