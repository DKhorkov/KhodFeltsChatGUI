package search_users_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	searchusers "github.com/DKhorkov/kfcGUI/internal/v2/handlers/search_users"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

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

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := searchusers.New(mockUseCases)

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

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := searchusers.New(mockUseCases)
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := searchusers.New(mockUseCases)
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := searchusers.New(mockUseCases)
	h.StopListening()
}
