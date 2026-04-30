package profile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	profilehandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/profile"
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

func TestHandler_ChangePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		in            domains.ChangePasswordDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name: "successful change password",
			in: domains.ChangePasswordDTO{
				OldPassword: "OldPassword1!",
				NewPassword: "NewPassword1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ChangePassword(gomock.Any(), domains.ChangePasswordDTO{
						OldPassword: "OldPassword1!",
						NewPassword: "NewPassword1!",
					}).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name: "invalid new password - too weak",
			in: domains.ChangePasswordDTO{
				OldPassword: "OldPassword1!",
				NewPassword: "weak",
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name: "invalid old password - too weak",
			in: domains.ChangePasswordDTO{
				OldPassword: "weak",
				NewPassword: "NewPassword1!",
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name: "invalid new password - no special character",
			in: domains.ChangePasswordDTO{
				OldPassword: "OldPassword1!",
				NewPassword: "NewPassword1",
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidPassword).
					Return(customerrors.ErrInvalidPassword).
					Times(1)
			},
			expectedError: customerrors.ErrInvalidPassword,
		},
		{
			name: "use case returns error",
			in: domains.ChangePasswordDTO{
				OldPassword: "OldPassword1!",
				NewPassword: "NewPassword1!",
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					ChangePassword(gomock.Any(), gomock.Any()).
					Return(errors.New("wrong password")).
					Times(1)
			},
			expectedError: errors.New("wrong password"),
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

			h := profilehandler.New(mockUseCases, mockMapper, testValidationConfig())

			err := h.ChangePassword(tt.in)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		in            domains.UpdateUserDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful update",
			in:   domains.UpdateUserDTO{Username: "newusername"},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					UpdateUser(gomock.Any(), domains.UpdateUserDTO{Username: "newusername"}).
					Return(&domains.User{ID: 1, Username: "newusername", Email: "john@example.com"}, nil).
					Times(1)
			},
			expectedUser:  &domains.User{ID: 1, Username: "newusername", Email: "john@example.com"},
			expectedError: nil,
		},
		{
			name: "invalid username - too short",
			in:   domains.UpdateUserDTO{Username: "ab"},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidUsername).
					Return(customerrors.ErrInvalidUsername).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: customerrors.ErrInvalidUsername,
		},
		{
			name: "invalid username - contains underscore",
			in:   domains.UpdateUserDTO{Username: "john_doe"},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidUsername).
					Return(customerrors.ErrInvalidUsername).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: customerrors.ErrInvalidUsername,
		},
		{
			name: "use case returns error",
			in:   domains.UpdateUserDTO{Username: "validname"},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("user not found")).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
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

			h := profilehandler.New(mockUseCases, mockMapper, testValidationConfig())

			user, err := h.UpdateUser(tt.in)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := profilehandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := profilehandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := profilehandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.StopListening()
}
