package errors

import (
	"errors"
	"testing"
)

func TestMapper_Map(t *testing.T) {
	t.Parallel()

	m := New()

	tests := []struct {
		name     string
		actual   error
		expected error
	}{
		{
			name:     "Nil error",
			actual:   nil,
			expected: nil,
		},
		{
			name:     "Known error(User not found)",
			actual:   errors.New("user not found"),
			expected: errUserNotFound,
		},
		{
			name:     "Known error(Chat not found)",
			actual:   errors.New("chat not found"),
			expected: errChatNotFound,
		},
		{
			name:     "Unknown error",
			actual:   errors.New("unknown error"),
			expected: errDefault,
		},
		{
			name:     "Similar but not matching",
			actual:   errors.New("user missing"),
			expected: errDefault,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := m.Map(test.actual)

			if test.expected == nil {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}

				return
			}

			if !errors.Is(result, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
