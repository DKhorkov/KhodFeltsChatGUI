package notifications

import "testing"

func TestFormatBadgeTitle(t *testing.T) {
	t.Parallel()

	const appTitle = "KFC Chat"

	tests := []struct {
		name  string
		total int
		want  string
	}{
		{name: "zero unread", total: 0, want: appTitle},
		{name: "negative falls back to plain title", total: -1, want: appTitle},
		{name: "single digit", total: 1, want: "(1) " + appTitle},
		{name: "two digits", total: 42, want: "(42) " + appTitle},
		{name: "boundary 99", total: 99, want: "(99) " + appTitle},
		{name: "over 99 clamped", total: 100, want: "(99+) " + appTitle},
		{name: "much over 99 clamped", total: 10_000, want: "(99+) " + appTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatBadgeTitle(tt.total, appTitle)
			if got != tt.want {
				t.Errorf("formatBadgeTitle(%d, %q) = %q, want %q", tt.total, appTitle, got, tt.want)
			}
		})
	}
}
