package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	settingshandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/settings"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupMocks       func(*mockusecases.MockUseCases)
		expectedSettings *domains.Settings
		expectedError    error
	}{
		{
			name: "successful get settings",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetSettings(gomock.Any()).
					Return(&domains.Settings{
						Theme:           domains.ThemeDark,
						EmailConsents:   domains.ConsentNewMessage,
						WebPushConsents: 0,
					}, nil).
					Times(1)
			},
			expectedSettings: &domains.Settings{
				Theme:           domains.ThemeDark,
				EmailConsents:   domains.ConsentNewMessage,
				WebPushConsents: 0,
			},
			expectedError: nil,
		},
		{
			name: "get settings error",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetSettings(gomock.Any()).
					Return(nil, errors.New("failed to get settings")).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("failed to get settings"),
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

			h := settingshandler.New(mockUseCases)

			settings, err := h.GetSettings()

			assert.Equal(t, tt.expectedSettings, settings)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_UpdateSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            domains.Settings
		setupMocks       func(*mockusecases.MockUseCases)
		expectedSettings *domains.Settings
		expectedError    error
	}{
		{
			name: "successful update settings",
			input: domains.Settings{
				Theme:           domains.ThemeDark,
				EmailConsents:   domains.ConsentNewMessage,
				WebPushConsents: domains.ConsentNewMessage,
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					UpdateSettings(gomock.Any(), domains.Settings{
						Theme:           domains.ThemeDark,
						EmailConsents:   domains.ConsentNewMessage,
						WebPushConsents: domains.ConsentNewMessage,
					}).
					Return(&domains.Settings{
						Theme:           domains.ThemeDark,
						EmailConsents:   domains.ConsentNewMessage,
						WebPushConsents: domains.ConsentNewMessage,
					}, nil).
					Times(1)
			},
			expectedSettings: &domains.Settings{
				Theme:           domains.ThemeDark,
				EmailConsents:   domains.ConsentNewMessage,
				WebPushConsents: domains.ConsentNewMessage,
			},
			expectedError: nil,
		},
		{
			name: "update settings error",
			input: domains.Settings{
				Theme:         domains.ThemeLight,
				EmailConsents: domains.ConsentNewMessage,
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					UpdateSettings(gomock.Any(), domains.Settings{
						Theme:         domains.ThemeLight,
						EmailConsents: domains.ConsentNewMessage,
					}).
					Return(nil, errors.New("update failed")).
					Times(1)
			},
			expectedSettings: nil,
			expectedError:    errors.New("update failed"),
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

			h := settingshandler.New(mockUseCases)

			settings, err := h.UpdateSettings(tt.input)

			assert.Equal(t, tt.expectedSettings, settings)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
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

	h := settingshandler.New(mockUseCases)
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := settingshandler.New(mockUseCases)
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := settingshandler.New(mockUseCases)
	h.StopListening()
}
