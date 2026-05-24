package users_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	usershandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/users"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/DKhorkov/libs/pointers"
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

func TestHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockusecases.MockUseCases)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful get current user",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1, Username: "john"}, nil).
					Times(1)
			},
			expectedUser:  &domains.User{ID: 1, Username: "john"},
			expectedError: nil,
		},
		{
			name: "use case error",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(nil, errors.New("session expired")).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("session expired"),
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

			h := usershandler.New(mockUseCases, mockMapper, testValidationConfig())

			user, err := h.GetCurrentUser()

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

func TestHandler_SearchUsers(t *testing.T) {
	t.Parallel()

	username := "john"
	now := time.Now()

	tests := []struct {
		name          string
		filters       *domains.UsersFilters
		pagination    *domains.Pagination
		setupMocks    func(*mockusecases.MockUseCases)
		expectedUsers []domains.User
		expectedError error
	}{
		{
			name:       "successful search with username filter",
			filters:    &domains.UsersFilters{Username: &username},
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					SearchUsers(
						gomock.Any(),
						&domains.UsersFilters{Username: &username},
						(*domains.Pagination)(nil),
					).
					Return([]domains.User{
						{
							ID:        1,
							Username:  "john",
							Email:     "john@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{
				{
					ID:        1,
					Username:  "john",
					Email:     "john@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
		{
			name:    "successful search with pagination",
			filters: &domains.UsersFilters{Username: &username},
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](5),
				Offset: pointers.New[uint64](0),
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					SearchUsers(
						gomock.Any(),
						&domains.UsersFilters{Username: &username},
						&domains.Pagination{
							Limit:  pointers.New[uint64](5),
							Offset: pointers.New[uint64](0),
						},
					).
					Return([]domains.User{{ID: 1, Username: "john"}}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{{ID: 1, Username: "john"}},
			expectedError: nil,
		},
		{
			name:       "no users found - empty result",
			filters:    &domains.UsersFilters{Username: &username},
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					SearchUsers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]domains.User{}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{},
			expectedError: nil,
		},
		{
			name:       "nil filters",
			filters:    nil,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					SearchUsers(gomock.Any(), (*domains.UsersFilters)(nil), (*domains.Pagination)(nil)).
					Return([]domains.User{{ID: 1}, {ID: 2}}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{{ID: 1}, {ID: 2}},
			expectedError: nil,
		},
		{
			name:       "use case error",
			filters:    &domains.UsersFilters{Username: &username},
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					SearchUsers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("database error"),
		},
		{
			name:    "multiple users found",
			filters: nil,
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](0),
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					SearchUsers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]domains.User{
						{ID: 1, Username: "alice"},
						{ID: 2, Username: "alex"},
						{ID: 3, Username: "alan"},
					}, nil).
					Times(1)
			},
			expectedUsers: []domains.User{
				{ID: 1, Username: "alice"},
				{ID: 2, Username: "alex"},
				{ID: 3, Username: "alan"},
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
				tt.setupMocks(mockUseCases)
			}

			h := usershandler.New(mockUseCases, mockMapper, testValidationConfig())

			users, err := h.SearchUsers(tt.filters, tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUsers, users)
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
			in:   domains.UpdateUserDTO{Username: pointers.New("newusername")},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					UpdateUser(gomock.Any(), domains.UpdateUserDTO{Username: pointers.New("newusername")}).
					Return(&domains.User{ID: 1, Username: "newusername", Email: "john@example.com"}, nil).
					Times(1)
			},
			expectedUser:  &domains.User{ID: 1, Username: "newusername", Email: "john@example.com"},
			expectedError: nil,
		},
		{
			name: "invalid username - too short",
			in:   domains.UpdateUserDTO{Username: pointers.New("ab")},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidLogin).
					Return(customerrors.ErrInvalidLogin).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: customerrors.ErrInvalidLogin,
		},
		{
			name: "invalid username - contains underscore",
			in:   domains.UpdateUserDTO{Username: pointers.New("john_doe")},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(customerrors.ErrInvalidLogin).
					Return(customerrors.ErrInvalidLogin).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: customerrors.ErrInvalidLogin,
		},
		{
			name: "use case returns error",
			in:   domains.UpdateUserDTO{Username: pointers.New("validname")},
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

			h := usershandler.New(mockUseCases, mockMapper, testValidationConfig())

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

	h := usershandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := usershandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := usershandler.New(mockUseCases, mockMapper, testValidationConfig())
	h.StopListening()
}
