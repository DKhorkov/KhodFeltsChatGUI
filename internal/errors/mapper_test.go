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
			expected: ErrDefault,
		},
		{
			name:     "Similar but not matching",
			actual:   errors.New("user missing"),
			expected: ErrDefault,
		},
		{
			name:     "Specific key matched before generic (new password equal to old)",
			actual:   errors.New("validation failed: new password can not be equal to old password"),
			expected: errNewPasswordEqualToOldPassword,
		},
		{
			name:     "Generic validation failed when no specific match",
			actual:   errors.New("validation failed"),
			expected: errValidationFailed,
		},
		{
			name:     "Substring match prefers longer key",
			actual:   errors.New("error: validation failed: new password can not be equal to old password"),
			expected: errNewPasswordEqualToOldPassword,
		},
		{
			name:     "Wrong password error",
			actual:   errors.New("wrong password"),
			expected: errLoginFailed,
		},
		{
			name:     "Error containing known key as substring",
			actual:   errors.New("something went wrong: login failed"),
			expected: errLoginFailed,
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
