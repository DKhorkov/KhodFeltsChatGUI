package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	authhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/auth"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func testValidationConfig() config.ValidationConfig {
	return config.ValidationConfig{
		EmailRegExp: `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`,
		PasswordRegExps: []string{
			`.{8,}`,
			`[a-z]`,
			`[A-Z]`,
			`[0-9]`,
			`[^\d\w]`,
		},
		UsernameRegExps: []string{
			`^.{5,70}$`,
			`^[A-Za-z0-9]+$`,
		},
	}
}

func TestHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		in            domains.LoginDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name: "successful login with email",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "Password1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "john@example.com", Password: "Password1!"}).
					Return(&domains.User{ID: 1, Username: "john"}, nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "successful login with username",
			in:   domains.LoginDTO{Login: "johnathon", Password: "Password1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "johnathon", Password: "Password1!"}).
					Return(&domains.User{ID: 1, Username: "johnathon"}, nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "invalid login",
			in:   domains.LoginDTO{Login: "inv", Password: "Password1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidLogin).
					Return(customerrors.ErrInvalidLogin).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidLogin,
		},
		{
			name: "invalid password - too weak",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "password"},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name: "invalid password - too short",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "Pass1!"},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name: "use case returns ErrDefault - mapped to ErrLogin",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "Password1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "john@example.com", Password: "Password1!"}).
					Return(nil, customerrors.ErrDefault).
					Times(1)
				em.EXPECT().Map(customerrors.ErrLogin).Return(customerrors.ErrLogin).Times(1)
			},
			expectedError: customerrors.ErrLogin,
		},
		{
			name: "use case returns network error - returned as-is",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "Password1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "john@example.com", Password: "Password1!"}).
					Return(nil, errors.New("connection refused")).
					Times(1)
			},
			expectedError: errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockMapper)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.Login(tt.in)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		in            domains.RegisterDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name: "successful registration",
			in: domains.RegisterDTO{
				Email:    "john@example.com",
				Username: "john123",
				Password: "Password1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Register(gomock.Any(), domains.RegisterDTO{
						Email:    "john@example.com",
						Username: "john123",
						Password: "Password1!",
					}).
					Return(&domains.User{ID: 1}, nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "invalid email",
			in: domains.RegisterDTO{
				Email:    "invalid-email",
				Username: "john123",
				Password: "Password1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidEmail).
					Return(customerrors.ErrInvalidEmail).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidEmail,
		},
		{
			name: "invalid username - too short",
			in: domains.RegisterDTO{
				Email:    "john@example.com",
				Username: "ab",
				Password: "Password1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidLogin).
					Return(customerrors.ErrInvalidLogin).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidLogin,
		},
		{
			name: "invalid username - contains underscore",
			in: domains.RegisterDTO{
				Email:    "john@example.com",
				Username: "john_doe",
				Password: "Password1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidLogin).
					Return(customerrors.ErrInvalidLogin).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidLogin,
		},
		{
			name: "invalid password - no special character",
			in: domains.RegisterDTO{
				Email:    "john@example.com",
				Username: "john123",
				Password: "Password1",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name: "use case returns ErrDefault - mapped to ErrRegister",
			in: domains.RegisterDTO{
				Email:    "john@example.com",
				Username: "john123",
				Password: "Password1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, customerrors.ErrDefault).
					Times(1)
				em.EXPECT().Map(customerrors.ErrRegister).Return(customerrors.ErrRegister).Times(1)
			},
			expectedError: customerrors.ErrRegister,
		},
		{
			name: "use case returns specific error - returned as-is",
			in: domains.RegisterDTO{
				Email:    "john@example.com",
				Username: "john123",
				Password: "Password1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, customerrors.ErrEmailAlreadyExists).
					Times(1)
			},
			expectedError: customerrors.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockMapper)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.Register(tt.in)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_SendVerifyEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name:  "successful send",
			email: "john@example.com",
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "john@example.com").
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "invalid email",
			email: "not-an-email",
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidEmail).
					Return(customerrors.ErrInvalidEmail).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidEmail,
		},
		{
			name:  "use case error",
			email: "john@example.com",
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "john@example.com").
					Return(errors.New("smtp unavailable")).
					Times(1)
			},
			expectedError: errors.New("smtp unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockMapper)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.SendVerifyEmail(tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_SendForgetPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name:  "successful send",
			email: "john@example.com",
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "john@example.com").
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "invalid email",
			email: "invalid",
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidEmail).
					Return(customerrors.ErrInvalidEmail).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidEmail,
		},
		{
			name:  "use case error",
			email: "john@example.com",
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "john@example.com").
					Return(customerrors.ErrUserNotFound).
					Times(1)
			},
			expectedError: customerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockMapper)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.SendForgetPassword(tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_ForgetPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		token         string
		in            domains.ForgetPasswordDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name:  "successful password reset",
			token: "valid-token",
			in:    domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "NewPassword1!").
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "invalid password",
			token: "valid-token",
			in:    domains.ForgetPasswordDTO{NewPassword: "weak"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name:  "use case error",
			token: "valid-token",
			in:    domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "NewPassword1!").
					Return(errors.New("token expired")).
					Times(1)
			},
			expectedError: errors.New("token expired"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockMapper)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.ForgetPassword(tt.token, tt.in)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_Authenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedError error
	}{
		{
			name: "authenticated successfully",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					Authenticate(gomock.Any()).
					Return(&domains.User{ID: 1}, nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "not authenticated",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					Authenticate(gomock.Any()).
					Return(nil, errors.New("session not found")).
					Times(1)
			},
			expectedError: errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())
			h.SetContext(context.Background())

			err := h.Authenticate()

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedError error
	}{
		{
			name: "successful logout",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					Logout(gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "logout error",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					Logout(gomock.Any()).
					Return(customerrors.ErrLogout).
					Times(1)
			},
			expectedError: customerrors.ErrLogout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.Logout()

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
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := authhandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.StopListening()
}
