package usecases_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	mockrepositories "github.com/DKhorkov/kfcGUI/mocks/repositories"
	"github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCases_Authenticate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockUsersRepository, *mockrepositories.MockAuthRepository, *mocks.MockLogger)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful authentication",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "new_access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "john_doe",
						Email:     "john@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedUser: &domains.User{
				ID:        1,
				Username:  "john_doe",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				_ *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to refresh tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(nil, errors.New("refresh failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to refresh tokens", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("refresh failed"),
		},
		{
			name: "failed to get current user",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "new_access_token").
					Return(nil, errors.New("user not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get user by refreshed tokens", gomock.Any()).
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

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockAuth, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			user, err := uc.Authenticate(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)
		})
	}
}

func TestUseCases_Login(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockUsersRepository, *mockrepositories.MockAuthRepository, *mocks.MockLogger)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name:     "successful login",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), "john@example.com", "password123").
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}).
					Return(nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "john_doe",
						Email:     "john@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedUser: &domains.User{
				ID:        1,
				Username:  "john_doe",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name:     "failed login",
			email:    "wrong@example.com",
			password: "wrong",
			setupMocks: func(
				_ *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), "wrong@example.com", "wrong").
					Return(nil, errors.New("invalid credentials"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to login", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("invalid credentials"),
		},
		{
			name:     "failed to save tokens",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), "john@example.com", "password123").
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}).
					Return(errors.New("save failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to save tokens", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("save failed"),
		},
		{
			name:     "failed to get current user after login",
			email:    "john@example.com",
			password: "password123",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), "john@example.com", "password123").
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}).
					Return(nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(nil, errors.New("user not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get user by accessToken", gomock.Any()).
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

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockAuth, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			user, err := uc.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)
		})
	}
}

func TestUseCases_Logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockAuthRepository, *mockrepositories.MockWebSocketsRepository, *mocks.MockLogger)
		expectedError error
	}{
		{
			name: "successful logout",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					Logout(gomock.Any(), "access_token").
					Return(nil)

				mockTokens.EXPECT().
					Delete(gomock.Any()).
					Return(nil)

				mockWS.EXPECT().
					Close().
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockAuthRepository,
				_ *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to logout from auth",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					Logout(gomock.Any(), "access_token").
					Return(errors.New("logout failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to logout", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("logout failed"),
		},
		{
			name: "failed to delete tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					Logout(gomock.Any(), "access_token").
					Return(nil)

				mockTokens.EXPECT().
					Delete(gomock.Any()).
					Return(errors.New("delete failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to delete tokens", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("delete failed"),
		},
		{
			name: "failed to close websocket",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					Logout(gomock.Any(), "access_token").
					Return(nil)

				mockTokens.EXPECT().
					Delete(gomock.Any()).
					Return(nil)

				mockWS.EXPECT().
					Close().
					Return(errors.New("close failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to close websockets", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("close failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockAuth, mockWS, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			err := uc.Logout(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_Register(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		registerData  domains.RegisterDTO
		setupMocks    func(*mockrepositories.MockAuthRepository, *mocks.MockLogger)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful registration",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(mockAuth *mockrepositories.MockAuthRepository, _ *mocks.MockLogger) {
				mockAuth.EXPECT().
					Register(gomock.Any(), domains.RegisterDTO{
						Username: "john_doe",
						Email:    "john@example.com",
						Password: "password123",
					}).
					Return(&domains.User{
						ID:        1,
						Username:  "john_doe",
						Email:     "john@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedUser: &domains.User{
				ID:        1,
				Username:  "john_doe",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name: "failed registration",
			registerData: domains.RegisterDTO{
				Username: "john_doe",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(mockAuth *mockrepositories.MockAuthRepository, mockLogger *mocks.MockLogger) {
				mockAuth.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("user already exists"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to register user", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("user already exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			user, err := uc.Register(context.Background(), tt.registerData)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)
		})
	}
}

func TestUseCases_GetUserChats(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		limit         int
		offset        int
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockChatsRepository, *mocks.MockLogger)
		expectedChats []domains.Chat
		expectedError error
	}{
		{
			name:   "successful get user chats",
			limit:  10,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					GetUserChats(gomock.Any(), "access_token", 10, 0).
					Return([]domains.Chat{
						{ID: 1, Title: pointers.New("Chat 1"), CreatedAt: now, UpdatedAt: now},
						{ID: 2, Title: pointers.New("Chat 2"), CreatedAt: now, UpdatedAt: now},
					}, nil)
			},
			expectedChats: []domains.Chat{
				{ID: 1, Title: pointers.New("Chat 1"), CreatedAt: now, UpdatedAt: now},
				{ID: 2, Title: pointers.New("Chat 2"), CreatedAt: now, UpdatedAt: now},
			},
			expectedError: nil,
		},
		{
			name:   "failed to load tokens",
			limit:  10,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedChats: nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name:   "failed to get user chats",
			limit:  10,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					GetUserChats(gomock.Any(), "access_token", 10, 0).
					Return(nil, errors.New("chats not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get user chats", gomock.Any()).
					Times(1)
			},
			expectedChats: nil,
			expectedError: errors.New("chats not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockChats, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			chats, err := uc.GetUserChats(context.Background(), tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedChats, chats)
		})
	}
}

func TestUseCases_SearchUsers(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		username      string
		limit         int
		offset        int
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockUsersRepository, *mocks.MockLogger)
		expectedUsers []domains.User
		expectedError error
	}{
		{
			name:     "successful search - filter out current user",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "current_user",
						Email:     "current@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				mockUsers.EXPECT().
					SearchUsers(gomock.Any(), "john", 10, 0).
					Return([]domains.User{
						{
							ID:        1,
							Username:  "current_user",
							Email:     "current@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							Username:  "john_doe",
							Email:     "john@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        3,
							Username:  "john_smith",
							Email:     "john.smith@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
			expectedUsers: []domains.User{
				{
					ID:        2,
					Username:  "john_doe",
					Email:     "john@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        3,
					Username:  "john_smith",
					Email:     "john.smith@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
		{
			name:     "search returns only other users",
			username: "other",
			limit:    10,
			offset:   0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "current_user",
						Email:     "current@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				mockUsers.EXPECT().
					SearchUsers(gomock.Any(), "other", 10, 0).
					Return([]domains.User{
						{
							ID:        2,
							Username:  "other_user",
							Email:     "other@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
			expectedUsers: []domains.User{
				{
					ID:        2,
					Username:  "other_user",
					Email:     "other@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedError: nil,
		},
		{
			name:     "empty search results",
			username: "nonexistent",
			limit:    10,
			offset:   0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "current_user",
						Email:     "current@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				mockUsers.EXPECT().
					SearchUsers(gomock.Any(), "nonexistent", 10, 0).
					Return([]domains.User{}, nil)
			},
			expectedUsers: []domains.User{},
			expectedError: nil,
		},
		{
			name:     "failed to load tokens",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name:     "failed to get current user",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(nil, errors.New("user not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get current user", gomock.Any()).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("user not found"),
		},
		{
			name:     "failed to search users",
			username: "john",
			limit:    10,
			offset:   0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "current_user",
						Email:     "current@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				mockUsers.EXPECT().
					SearchUsers(gomock.Any(), "john", 10, 0).
					Return(nil, errors.New("search failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to search users", gomock.Any()).
					Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("search failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			users, err := uc.SearchUsers(context.Background(), tt.username, tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUsers, users)
		})
	}
}

func TestUseCases_GetChatMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		chatID           uint64
		limit            int
		offset           int
		setupMocks       func(*mockrepositories.MockTokensRepository, *mockrepositories.MockChatsRepository, *mocks.MockLogger)
		expectedMessages []domains.Message
		expectedError    error
	}{
		{
			name:   "successful get chat messages with timezone conversion",
			chatID: 1,
			limit:  20,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				utcTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				mockChats.EXPECT().
					GetChatMessages(gomock.Any(), "access_token", uint64(1), 20, 0).
					Return([]domains.Message{
						{
							ID:        1,
							ChatID:    1,
							Text:      "Hello",
							CreatedAt: utcTime,
							UpdatedAt: utcTime,
						},
					}, nil)
			},
			expectedMessages: []domains.Message{
				{
					ID:        1,
					ChatID:    1,
					Text:      "Hello",
					CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, common.Timezone),
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, common.Timezone),
				},
			},
			expectedError: nil,
		},
		{
			name:   "failed to load tokens",
			chatID: 1,
			limit:  10,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New("tokens not found"),
		},
		{
			name:   "failed to get chat messages",
			chatID: 1,
			limit:  10,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					GetChatMessages(gomock.Any(), "access_token", uint64(1), 10, 0).
					Return(nil, errors.New("messages not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get chat messages", gomock.Any()).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New("messages not found"),
		},
		{
			name:   "empty messages list",
			chatID: 1,
			limit:  10,
			offset: 0,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					GetChatMessages(gomock.Any(), "access_token", uint64(1), 10, 0).
					Return([]domains.Message{}, nil)
			},
			expectedMessages: []domains.Message{},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockChats, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			messages, err := uc.GetChatMessages(
				context.Background(),
				tt.chatID,
				tt.limit,
				tt.offset,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			reflect.DeepEqual(tt.expectedMessages, messages)
		})
	}
}

func TestUseCases_SendMessage(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		message       domains.Message
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockWebSocketsRepository, *mocks.MockLogger)
		expectedError error
	}{
		{
			name: "successful send message",
			message: domains.Message{
				ID:     1,
				ChatID: 1,
				Sender: domains.User{
					ID:        1,
					Username:  "john_doe",
					Email:     "john@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Text:      "Hello, world!",
				CreatedAt: now,
				UpdatedAt: now,
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					WriteMessage(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, msg domains.Message) error {
						assert.Equal(t, "Hello, world!", msg.Text)

						return nil
					})
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			message: domains.Message{
				Text: "Hello",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to connect to websocket",
			message: domains.Message{
				Text: "Hello",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(errors.New("connection failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to connect to websockets", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("connection failed"),
		},
		{
			name: "failed to write message",
			message: domains.Message{
				Text: "Hello",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					WriteMessage(gomock.Any(), gomock.Any()).
					Return(errors.New("write failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to send message", gomock.Any()).
					Times(1)
			},
			expectedError: errors.New("write failed"),
		},
		{
			name: "send empty message",
			message: domains.Message{
				Text: "",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					WriteMessage(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, msg domains.Message) error {
						assert.Empty(t, msg.Text)

						return nil
					})
			},
			expectedError: nil,
		},
		{
			name: "send message with special characters",
			message: domains.Message{
				Text: "Hello! @#$%^&*()",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					WriteMessage(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, msg domains.Message) error {
						assert.Equal(t, "Hello! @#$%^&*()", msg.Text)

						return nil
					})
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockWS, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			err := uc.SendMessage(context.Background(), tt.message)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_ReadMessage(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name            string
		setupMocks      func(*mockrepositories.MockTokensRepository, *mockrepositories.MockWebSocketsRepository, *mocks.MockLogger)
		expectedMessage *domains.Message
		expectedError   error
	}{
		{
			name: "successful read message",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					ReadMessage(gomock.Any()).
					Return(&domains.Message{
						ID:     1,
						ChatID: 1,
						Sender: domains.User{
							ID:        1,
							Username:  "john_doe",
							Email:     "john@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
						Text:      "Hello, world!",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedMessage: &domains.Message{
				ID:     1,
				ChatID: 1,
				Sender: domains.User{
					ID:        1,
					Username:  "john_doe",
					Email:     "john@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Text:      "Hello, world!",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("tokens not found"),
		},
		{
			name: "failed to connect to websocket",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(errors.New("connection failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to connect to websockets", gomock.Any()).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("connection failed"),
		},
		{
			name: "failed to read message",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					ReadMessage(gomock.Any()).
					Return(nil, errors.New("read failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to read message", gomock.Any()).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("read failed"),
		},
		{
			name: "read empty message",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockWS.EXPECT().
					Connect(gomock.Any(), "access_token").
					Return(nil)

				mockWS.EXPECT().
					ReadMessage(gomock.Any()).
					Return(&domains.Message{
						Text: "",
					}, nil)
			},
			expectedMessage: &domains.Message{
				Text: "",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockWS, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			message, err := uc.ReadMessage(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedMessage, message)
		})
	}
}

func TestUseCases_CreateChat(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		chat          domains.Chat
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockChatsRepository, *mocks.MockLogger)
		expectedChat  *domains.Chat
		expectedError error
	}{
		{
			name: "successful create chat",
			chat: domains.Chat{
				Title:     pointers.New("New Chat"),
				CreatedAt: now,
				UpdatedAt: now,
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					CreateChat(gomock.Any(), "access_token", domains.Chat{
						Title:     pointers.New("New Chat"),
						CreatedAt: now,
						UpdatedAt: now,
					}).
					Return(&domains.Chat{
						ID:        1,
						Title:     pointers.New("New Chat"),
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedChat: &domains.Chat{
				ID:        1,
				Title:     pointers.New("New Chat"),
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name: "create chat with empty name",
			chat: domains.Chat{
				Title:     pointers.New(""),
				CreatedAt: now,
				UpdatedAt: now,
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					CreateChat(gomock.Any(), "access_token", gomock.Any()).
					Return(nil, errors.New("chat name is required"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to create chat", gomock.Any()).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: errors.New("chat name is required"),
		},
		{
			name: "failed to load tokens",
			chat: domains.Chat{
				Title: pointers.New("New Chat"),
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "chat already exists",
			chat: domains.Chat{
				Title: pointers.New("Existing Chat"),
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					CreateChat(gomock.Any(), "access_token", gomock.Any()).
					Return(nil, errors.New("chat already exists"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to create chat", gomock.Any()).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: errors.New("chat already exists"),
		},
		{
			name: "create chat with long name",
			chat: domains.Chat{
				Title: pointers.New(
					"This is a very long chat name that exceeds normal length requirements for testing purposes",
				),
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					CreateChat(gomock.Any(), "access_token", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, chat domains.Chat) (*domains.Chat, error) {
						assert.Equal(
							t,
							pointers.New(
								"This is a very long chat name that exceeds normal length requirements for testing purposes",
							),
							chat.Title,
						)

						return &domains.Chat{
							ID:    1,
							Title: chat.Title,
						}, nil
					})
			},
			expectedChat: &domains.Chat{
				ID: 1,
				Title: pointers.New(
					"This is a very long chat name that exceeds normal length requirements for testing purposes",
				),
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockChats, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			createdChat, err := uc.CreateChat(context.Background(), tt.chat)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			reflect.DeepEqual(tt.expectedChat, createdChat)
		})
	}
}

func TestUseCases_GetCurrentUser(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockUsersRepository, *mockrepositories.MockAuthRepository, *mocks.MockLogger)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful get current user",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "new_access_token").
					Return(&domains.User{
						ID:        1,
						Username:  "john_doe",
						Email:     "john@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedUser: &domains.User{
				ID:        1,
				Username:  "john_doe",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: nil,
		},
		{
			name: "failed to refresh tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(nil, errors.New("refresh failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to refresh tokens", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("refresh failed"),
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				_ *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to get user after token refresh",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(nil)

				mockUsers.EXPECT().
					GetCurrentUser(gomock.Any(), "new_access_token").
					Return(nil, errors.New("user not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get user by refreshed tokens", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
		},
		{
			name: "failed to save refreshed tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(errors.New("save failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to save tokens", gomock.Any()).
					Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("save failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockAuth, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			user, err := uc.GetCurrentUser(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)
		})
	}
}

func TestUseCases_RefreshTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMocks     func(*mockrepositories.MockTokensRepository, *mockrepositories.MockAuthRepository, *mocks.MockLogger)
		expectedTokens *domains.TokensDTO
		expectedError  error
	}{
		{
			name: "successful refresh tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "old_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "old_refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(nil)
			},
			expectedTokens: &domains.TokensDTO{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens file not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("tokens file not found"),
		},
		{
			name: "failed to refresh tokens - invalid refresh token",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "expired_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "expired_refresh_token").
					Return(nil, errors.New("refresh token expired"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to refresh tokens", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("refresh token expired"),
		},
		{
			name: "failed to refresh tokens - network error",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token").
					Return(nil, errors.New("connection timeout"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to refresh tokens", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("connection timeout"),
		},
		{
			name: "failed to save refreshed tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "old_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "old_refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(errors.New("failed to write to file"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to save tokens", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("failed to write to file"),
		},
		{
			name: "load returns empty tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "",
						RefreshToken: "",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "").
					Return(nil, errors.New("refresh token is empty"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to refresh tokens", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("refresh token is empty"),
		},
		{
			name: "refresh returns empty tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "old_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "old_refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "",
						RefreshToken: "",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "",
						RefreshToken: "",
					}).
					Return(nil)
			},
			expectedTokens: &domains.TokensDTO{
				AccessToken:  "",
				RefreshToken: "",
			},
			expectedError: nil,
		},
		{
			name: "refresh tokens with context cancellation",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "old_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "old_refresh_token").
					Return(nil, context.Canceled)

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to refresh tokens", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  context.Canceled,
		},
		{
			name: "refresh tokens with multiple attempts",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "old_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "old_refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token_v2",
						RefreshToken: "new_refresh_token_v2",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token_v2",
						RefreshToken: "new_refresh_token_v2",
					}).
					Return(nil)
			},
			expectedTokens: &domains.TokensDTO{
				AccessToken:  "new_access_token_v2",
				RefreshToken: "new_refresh_token_v2",
			},
			expectedError: nil,
		},
		{
			name: "load tokens with invalid data",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("invalid JSON format"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("invalid JSON format"),
		},
		{
			name: "save tokens with permission denied",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "old_access_token",
						RefreshToken: "old_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "old_refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}).
					Return(errors.New("permission denied"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to save tokens", gomock.Any()).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("permission denied"),
		},
		{
			name: "refresh tokens with same tokens returned",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "same_access_token",
						RefreshToken: "same_refresh_token",
					}, nil)

				mockAuth.EXPECT().
					RefreshTokens(gomock.Any(), "same_refresh_token").
					Return(&domains.TokensDTO{
						AccessToken:  "same_access_token",
						RefreshToken: "same_refresh_token",
					}, nil)

				mockTokens.EXPECT().
					Save(gomock.Any(), domains.TokensDTO{
						AccessToken:  "same_access_token",
						RefreshToken: "same_refresh_token",
					}).
					Return(nil)
			},
			expectedTokens: &domains.TokensDTO{
				AccessToken:  "same_access_token",
				RefreshToken: "same_refresh_token",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockAuth, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			tokens, err := uc.RefreshTokens(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedTokens, tokens)
		})
	}
}

func TestUseCases_GetTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(*mockrepositories.MockSettingsRepository, *mocks.MockLogger)
		expected   domains.ThemeType
	}{
		{
			name: "successful get light theme",
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeLight,
					}, nil)
			},
			expected: domains.ThemeLight,
		},
		{
			name: "successful get dark theme",
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeDark,
					}, nil)
			},
			expected: domains.ThemeDark,
		},
		{
			name: "load error - return default light theme",
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("settings file not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load settings", gomock.Any()).
					Times(1)
			},
			expected: domains.ThemeLight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockSettings, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			ctx := context.Background()
			result := uc.GetTheme(ctx)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUseCases_SetTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		theme         domains.ThemeType
		setupMocks    func(*mockrepositories.MockSettingsRepository, *mocks.MockLogger)
		expectedError error
	}{
		{
			name:  "successful set light theme when settings exist",
			theme: domains.ThemeLight,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeDark,
					}, nil)

				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeLight,
					}).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "successful set dark theme when settings exist",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeLight,
					}, nil)

				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeDark,
					}).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "successful set theme when settings do not exist - create new settings",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("settings not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load settings", gomock.Any()).
					Times(1)

				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeDark,
					}).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "failed to save after loading existing settings",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeLight,
					}, nil)

				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeDark,
					}).
					Return(errors.New("save failed"))
			},
			expectedError: errors.New("save failed"),
		},
		{
			name:  "failed to save when creating new settings",
			theme: domains.ThemeLight,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("settings not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load settings", gomock.Any()).
					Times(1)

				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeLight,
					}).
					Return(errors.New("save failed"))
			},
			expectedError: errors.New("save failed"),
		},
		{
			name:  "set same theme - should update",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeDark,
					}, nil)

				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeDark,
					}).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "set theme after multiple operations",
			theme: domains.ThemeLight,
			setupMocks: func(
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				// Первая загрузка
				mockSettings.EXPECT().
					Load(gomock.Any()).
					Return(&domains.Settings{
						Theme: domains.ThemeDark,
					}, nil)

				// Сохранение
				mockSettings.EXPECT().
					Save(gomock.Any(), domains.Settings{
						Theme: domains.ThemeLight,
					}).
					Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
			mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
			mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
			mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
			mockChats := mockrepositories.NewMockChatsRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockSettings, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
			)

			ctx := context.Background()
			err := uc.SetTheme(ctx, tt.theme)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
