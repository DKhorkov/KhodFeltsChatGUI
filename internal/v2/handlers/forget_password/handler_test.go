package forget_password_test

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	forgetpassword "github.com/DKhorkov/kfcGUI/internal/v2/handlers/forget_password"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// validForgetPasswordToken — корректный токен: base64-raw кодировка числового ID пользователя.
// security.RawDecode декодирует его обратно в "42", которое парсится как uint64.
var validForgetPasswordToken = base64.RawURLEncoding.EncodeToString([]byte("42"))

func testValidationConfig() config.ValidationConfig {
	return config.ValidationConfig{
		PasswordRegExps: []string{
			`.{8,}`,
			`[a-z]`,
			`[A-Z]`,
			`[0-9]`,
			`[^\d\w]`,
		},
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
			token: validForgetPasswordToken,
			in:    domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ForgetPassword(gomock.Any(), validForgetPasswordToken, "NewPassword1!").
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:  "invalid password - too weak",
			token: validForgetPasswordToken,
			in:    domains.ForgetPasswordDTO{NewPassword: "weak"},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name:  "invalid password - no special character",
			token: validForgetPasswordToken,
			in:    domains.ForgetPasswordDTO{NewPassword: "Password1"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name:  "token expired - use case error",
			token: validForgetPasswordToken,
			in:    domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ForgetPassword(gomock.Any(), validForgetPasswordToken, "NewPassword1!").
					Return(errors.New("token has expired")).
					Times(1)
			},
			expectedError: errors.New("token has expired"),
		},
		{
			name:  "user not found - use case error",
			token: validForgetPasswordToken,
			in:    domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ForgetPassword(gomock.Any(), validForgetPasswordToken, "NewPassword1!").
					Return(customerrors.ErrUserNotFound).
					Times(1)
			},
			expectedError: customerrors.ErrUserNotFound,
		},
		{
			name:  "empty token",
			token: "",
			in:    domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidForgetPasswordToken).
					Return(customerrors.ErrInvalidForgetPasswordToken).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidForgetPasswordToken,
		},
		{
			name: "valid token for large user ID",
			// base64 raw url encoding of "9999999999"
			token: base64.RawURLEncoding.EncodeToString(
				[]byte(strconv.FormatUint(uint64(9999999999), 10)),
			),
			in: domains.ForgetPasswordDTO{NewPassword: "NewPassword1!"},
			setupMocks: func(uc *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				largeIDToken := base64.RawURLEncoding.EncodeToString([]byte("9999999999"))
				uc.EXPECT().
					ForgetPassword(gomock.Any(), largeIDToken, "NewPassword1!").
					Return(nil).
					Times(1)
			},
			expectedError: nil,
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

			h := forgetpassword.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.ForgetPassword(tt.token, tt.in)

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

	h := forgetpassword.New(mockUseCases, mockMapper, testValidationConfig())
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := forgetpassword.New(mockUseCases, mockMapper, testValidationConfig())
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := forgetpassword.New(mockUseCases, mockMapper, testValidationConfig())
	h.StopListening()
}
