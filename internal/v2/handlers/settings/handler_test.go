package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	settings "github.com/DKhorkov/kfcGUI/internal/v2/handlers/settings"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedTheme domains.ThemeType
	}{
		{
			name: "returns light theme",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeLight).
					Times(1)
			},
			expectedTheme: domains.ThemeLight,
		},
		{
			name: "returns dark theme",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeDark).
					Times(1)
			},
			expectedTheme: domains.ThemeDark,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := settings.New(mockUseCases)

			theme := h.GetTheme()

			assert.Equal(t, tt.expectedTheme, theme)
		})
	}
}

func TestHandler_ToggleTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedTheme domains.ThemeType
		expectedError error
	}{
		{
			name: "toggle from light to dark",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeLight).
					Times(1)
				uc.EXPECT().
					SetTheme(gomock.Any(), domains.ThemeDark).
					Return(nil).
					Times(1)
			},
			expectedTheme: domains.ThemeDark,
			expectedError: nil,
		},
		{
			name: "toggle from dark to light",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeDark).
					Times(1)
				uc.EXPECT().
					SetTheme(gomock.Any(), domains.ThemeLight).
					Return(nil).
					Times(1)
			},
			expectedTheme: domains.ThemeLight,
			expectedError: nil,
		},
		{
			name: "set theme error returns ThemeLight",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetTheme(gomock.Any()).
					Return(domains.ThemeLight).
					Times(1)
				uc.EXPECT().
					SetTheme(gomock.Any(), domains.ThemeDark).
					Return(errors.New("storage error")).
					Times(1)
			},
			expectedTheme: domains.ThemeLight,
			expectedError: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := settings.New(mockUseCases)

			theme, err := h.ToggleTheme()

			assert.Equal(t, tt.expectedTheme, theme)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := settings.New(mockUseCases)
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := settings.New(mockUseCases)
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := settings.New(mockUseCases)
	h.StopListening()
}
