package usecases_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
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
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mockrepositories.MockAuthRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
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
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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
		name       string
		in         domains.LoginDTO
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mockrepositories.MockAuthRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name: "successful login",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "password123"},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "john@example.com", Password: "password123"}).
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
			name: "failed login",
			in:   domains.LoginDTO{Login: "wrong@example.com", Password: "wrong"},
			setupMocks: func(
				_ *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "wrong@example.com", Password: "wrong"}).
					Return(nil, errors.New("invalid credentials"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to login", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("invalid credentials"),
		},
		{
			name: "failed to save tokens",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "password123"},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "john@example.com", Password: "password123"}).
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("save failed"),
		},
		{
			name: "failed to get current user after login",
			in:   domains.LoginDTO{Login: "john@example.com", Password: "password123"},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					Login(gomock.Any(), domains.LoginDTO{Login: "john@example.com", Password: "password123"}).
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			user, err := uc.Login(context.Background(), tt.in)

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
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockAuthRepository,
			*mockrepositories.MockWebSocketsRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError error
	}{
		{
			name: "successful logout",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockAuth, mockWS, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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
		name         string
		registerData domains.RegisterDTO
		setupMocks   func(
			*mockrepositories.MockAuthRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
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
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
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
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("user already exists"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to register user", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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

	pagination := &domains.Pagination{
		Limit:  pointers.New[uint64](10),
		Offset: pointers.New[uint64](10),
	}

	tests := []struct {
		name       string
		pagination *domains.Pagination
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockChatsRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedChats []domains.Chat
		expectedError error
	}{
		{
			name:       "successful get user chats",
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					GetUserChats(gomock.Any(), "access_token", pagination).
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
			name:       "failed to load tokens",
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedChats: nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name:       "failed to get user chats",
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockChats *mockrepositories.MockChatsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockChats.EXPECT().
					GetUserChats(gomock.Any(), "access_token", pagination).
					Return(nil, errors.New("chats not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get user chats", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockChats, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			chats, err := uc.GetUserChats(context.Background(), tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedChats, chats)
		})
	}
}

func TestUseCases_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		updateUserData domains.UpdateUserDTO
		setupMocks     func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedUser  *domains.User
		expectedError error
	}{
		{
			name:           "successful update user",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					UpdateUser(gomock.Any(), "access_token", domains.UpdateUserDTO{Username: pointers.New("newusername")}).
					Return(&domains.User{ID: 1, Username: "newusername", Email: "john@example.com"}, nil)
			},
			expectedUser:  &domains.User{ID: 1, Username: "newusername", Email: "john@example.com"},
			expectedError: nil,
		},
		{
			name:           "failed to load tokens",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedUser:  nil,
			expectedError: errors.New("test"),
		},
		{
			name:           "failed to update user",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("user not found")
				mappedErr := errors.New("user not found")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					UpdateUser(gomock.Any(), "access_token", gomock.Any()).
					Return(nil, originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to update user",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
		},
		{
			name:           "errors mapper transforms error",
			updateUserData: domains.UpdateUserDTO{Username: pointers.New("newusername")},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("some internal error")
				mappedErr := errors.New("user-friendly error message")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					UpdateUser(gomock.Any(), "access_token", gomock.Any()).
					Return(nil, originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to update user",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedUser:  nil,
			expectedError: errors.New("user-friendly error message"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			user, err := uc.UpdateUser(ctx, tt.updateUserData)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestUseCases_UpdateAvatar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fileData   []byte
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedURL   string
		expectedError error
	}{
		{
			name:     "successful avatar upload",
			fileData: []byte("fake image data"),
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					UpdateAvatar(gomock.Any(), "access_token", []byte("fake image data")).
					Return("https://kfc.webtm.ru/api/files/download/uuid.jpg", nil)
			},
			expectedURL:   "https://kfc.webtm.ru/api/files/download/uuid.jpg",
			expectedError: nil,
		},
		{
			name:     "failed to load tokens",
			fileData: []byte("data"),
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedURL:   "",
			expectedError: errors.New("test"),
		},
		{
			name:     "failed to update avatar",
			fileData: []byte("data"),
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("invalid image format")
				mappedErr := errors.New("invalid image format")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					UpdateAvatar(gomock.Any(), "access_token", gomock.Any()).
					Return("", originalErr)

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to update avatar", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedURL:   "",
			expectedError: errors.New("invalid image format"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			url, err := uc.UpdateAvatar(ctx, tt.fileData)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestUseCases_DeleteAvatar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError error
	}{
		{
			name: "successful avatar delete",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					DeleteAvatar(gomock.Any(), "access_token").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedError: errors.New("test"),
		},
		{
			name: "failed to delete avatar",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("user not found")
				mappedErr := errors.New("user not found")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockUsers.EXPECT().
					DeleteAvatar(gomock.Any(), "access_token").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to delete avatar", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			err := uc.DeleteAvatar(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_SearchUsers(t *testing.T) {
	t.Parallel()

	now := time.Now()

	filters := &domains.UsersFilters{
		Username: pointers.New("john"),
	}

	pagination := &domains.Pagination{
		Limit:  pointers.New[uint64](10),
		Offset: pointers.New[uint64](10),
	}

	tests := []struct {
		name       string
		filters    *domains.UsersFilters
		pagination *domains.Pagination
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedUsers []domains.User
		expectedError error
	}{
		{
			name:       "successful search - filter out current user",
			filters:    filters,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
					SearchUsers(gomock.Any(), filters, pagination).
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
			name:       "search returns only other users",
			filters:    filters,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
					SearchUsers(gomock.Any(), filters, pagination).
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
			name:       "empty search results",
			filters:    filters,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
					SearchUsers(gomock.Any(), filters, pagination).
					Return([]domains.User{}, nil)
			},
			expectedUsers: []domains.User{},
			expectedError: nil,
		},
		{
			name:       "failed to load tokens",
			filters:    filters,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name:       "failed to get current user",
			filters:    filters,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedUsers: nil,
			expectedError: errors.New("user not found"),
		},
		{
			name:       "failed to search users",
			filters:    filters,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockUsers *mockrepositories.MockUsersRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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
					SearchUsers(gomock.Any(), filters, pagination).
					Return(nil, errors.New("search failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to search users", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			users, err := uc.SearchUsers(context.Background(), tt.filters, tt.pagination)

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

	pagination := &domains.Pagination{
		Limit:  pointers.New[uint64](10),
		Offset: pointers.New[uint64](10),
	}

	tests := []struct {
		name       string
		chatID     uint64
		pagination *domains.Pagination
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedMessages []domains.Message
		expectedError    error
	}{
		{
			name:       "successful get chat messages with timezone conversion",
			chatID:     1,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				utcTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				mockMessages.EXPECT().
					GetChatMessages(gomock.Any(), "access_token", uint64(1), pagination).
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
			name:       "failed to load tokens",
			chatID:     1,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New("tokens not found"),
		},
		{
			name:       "failed to get chat messages",
			chatID:     1,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					GetChatMessages(gomock.Any(), "access_token", uint64(1), pagination).
					Return(nil, errors.New("messages not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get chat messages", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedMessages: nil,
			expectedError:    errors.New("messages not found"),
		},
		{
			name:       "empty messages list",
			chatID:     1,
			pagination: pagination,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					GetChatMessages(gomock.Any(), "access_token", uint64(1), pagination).
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			messages, err := uc.GetChatMessages(
				context.Background(),
				tt.chatID,
				tt.pagination,
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
		name       string
		message    domains.Message
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockWebSocketsRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
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
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				_ *mockerrors.MockErrorsMapper,
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
				_ *mockerrors.MockErrorsMapper,
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockWS, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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

func TestUseCases_ReadEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockWebSocketsRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper)
		expectedEvent *domains.WSEvent
		expectedError error
	}{
		{
			name: "successful read event",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
					ReadEvent(gomock.Any()).
					Return(&domains.WSEvent{
						Type:    domains.WSEventNewMessage,
						Payload: json.RawMessage(`{"id":1}`),
					}, nil)
			},
			expectedEvent: &domains.WSEvent{
				Type:    domains.WSEventNewMessage,
				Payload: json.RawMessage(`{"id":1}`),
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedEvent: nil,
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to connect to websocket",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedEvent: nil,
			expectedError: errors.New("connection failed"),
		},
		{
			name: "failed to read event",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockWS *mockrepositories.MockWebSocketsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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
					ReadEvent(gomock.Any()).
					Return(nil, errors.New("read failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to read event", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedEvent: nil,
			expectedError: errors.New("read failed"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockWS, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			event, err := uc.ReadEvent(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedEvent, event)
		})
	}
}

func TestUseCases_CreateChat(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		chat       domains.Chat
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockChatsRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper)
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
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				_ *mockerrors.MockErrorsMapper,
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockChats, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockUsersRepository,
			*mockrepositories.MockAuthRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
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
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockUsers, mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockAuthRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedTokens *domains.TokensDTO
		expectedError  error
	}{
		{
			name: "successful refresh tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens file not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "failed to refresh tokens - invalid refresh token",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "failed to refresh tokens - network error",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "failed to save refreshed tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "load returns empty tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "refresh returns empty tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(context.Canceled).
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
				_ *mockerrors.MockErrorsMapper,
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
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("invalid JSON format"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "save tokens with permission denied",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
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

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).Return(errors.New("test")).
					Times(1)
			},
			expectedTokens: nil,
			expectedError:  errors.New("test"),
		},
		{
			name: "refresh tokens with same tokens returned",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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
		setupMocks func(*mockrepositories.MockTokensRepository, *mockrepositories.MockSettingsRepository, *mocks.MockLogger)
		expected   domains.ThemeType
	}{
		{
			name: "successful get light theme",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(&domains.Settings{
						Theme: domains.ThemeLight,
					}, nil)
			},
			expected: domains.ThemeLight,
		},
		{
			name: "successful get dark theme",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(&domains.Settings{
						Theme: domains.ThemeDark,
					}, nil)
			},
			expected: domains.ThemeDark,
		},
		{
			name: "tokens load error - return default light theme",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)
			},
			expected: domains.ThemeLight,
		},
		{
			name: "get settings error - return default light theme",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(nil, errors.New("settings not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get settings", gomock.Any()).
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockSettings, mockLogger)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			result := uc.GetTheme(ctx)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUseCases_GetSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupMocks       func(*mockrepositories.MockTokensRepository, *mockrepositories.MockSettingsRepository, *mocks.MockLogger, *mockerrors.MockErrorsMapper)
		expectedSettings *domains.Settings
		expectedError    error
	}{
		{
			name: "successful get settings",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(&domains.Settings{
						Theme:           domains.ThemeDark,
						EmailConsents:   domains.ConsentNewMessage,
						WebPushConsents: 0,
					}, nil)
			},
			expectedSettings: &domains.Settings{
				Theme:           domains.ThemeDark,
				EmailConsents:   domains.ConsentNewMessage,
				WebPushConsents: 0,
			},
			expectedError: nil,
		},
		{
			name: "tokens load error",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("tokens not found"))
			},
			expectedSettings: nil,
			expectedError:    errors.New("tokens not found"),
		},
		{
			name: "get settings error",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(nil, errors.New("settings error"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get settings", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("settings error"))
			},
			expectedSettings: nil,
			expectedError:    errors.New("settings error"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockSettings, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			result, err := uc.GetSettings(ctx)

			assert.Equal(t, tt.expectedSettings, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_UpdateSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            domains.Settings
		setupMocks       func(*mockrepositories.MockTokensRepository, *mockrepositories.MockSettingsRepository, *mocks.MockLogger, *mockerrors.MockErrorsMapper)
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
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					UpdateSettings(gomock.Any(), "token", domains.Settings{
						Theme:           domains.ThemeDark,
						EmailConsents:   domains.ConsentNewMessage,
						WebPushConsents: domains.ConsentNewMessage,
					}).
					Return(&domains.Settings{
						Theme:           domains.ThemeDark,
						EmailConsents:   domains.ConsentNewMessage,
						WebPushConsents: domains.ConsentNewMessage,
					}, nil)
			},
			expectedSettings: &domains.Settings{
				Theme:           domains.ThemeDark,
				EmailConsents:   domains.ConsentNewMessage,
				WebPushConsents: domains.ConsentNewMessage,
			},
			expectedError: nil,
		},
		{
			name: "tokens load error",
			input: domains.Settings{
				Theme:         domains.ThemeLight,
				EmailConsents: domains.ConsentNewMessage,
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("tokens not found"))
			},
			expectedSettings: nil,
			expectedError:    errors.New("tokens not found"),
		},
		{
			name: "update settings error",
			input: domains.Settings{
				Theme:           domains.ThemeDark,
				WebPushConsents: domains.ConsentNewMessage,
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					UpdateSettings(gomock.Any(), "token", domains.Settings{
						Theme:           domains.ThemeDark,
						WebPushConsents: domains.ConsentNewMessage,
					}).
					Return(nil, errors.New("update failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to update settings", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("update failed"))
			},
			expectedSettings: nil,
			expectedError:    errors.New("update failed"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockSettings, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			result, err := uc.UpdateSettings(ctx, tt.input)

			assert.Equal(t, tt.expectedSettings, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_SetTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		theme         domains.ThemeType
		setupMocks    func(*mockrepositories.MockTokensRepository, *mockrepositories.MockSettingsRepository, *mocks.MockLogger, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name:  "successful set light theme",
			theme: domains.ThemeLight,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(&domains.Settings{Theme: domains.ThemeDark, EmailConsents: 1}, nil)

				mockSettings.EXPECT().
					UpdateSettings(gomock.Any(), "token", domains.Settings{
						Theme:         domains.ThemeLight,
						EmailConsents: 1,
					}).
					Return(&domains.Settings{Theme: domains.ThemeLight, EmailConsents: 1}, nil)
			},
			expectedError: nil,
		},
		{
			name:  "successful set dark theme",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(&domains.Settings{Theme: domains.ThemeLight}, nil)

				mockSettings.EXPECT().
					UpdateSettings(gomock.Any(), "token", domains.Settings{
						Theme: domains.ThemeDark,
					}).
					Return(&domains.Settings{Theme: domains.ThemeDark}, nil)
			},
			expectedError: nil,
		},
		{
			name:  "tokens load error",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("tokens not found"))
			},
			expectedError: errors.New("tokens not found"),
		},
		{
			name:  "get settings error",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(nil, errors.New("get settings failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get settings", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("get settings failed"))
			},
			expectedError: errors.New("get settings failed"),
		},
		{
			name:  "update settings error",
			theme: domains.ThemeDark,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockSettings *mockrepositories.MockSettingsRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "token", RefreshToken: "refresh"}, nil)

				mockSettings.EXPECT().
					GetSettings(gomock.Any(), "token").
					Return(&domains.Settings{Theme: domains.ThemeLight}, nil)

				mockSettings.EXPECT().
					UpdateSettings(gomock.Any(), "token", domains.Settings{
						Theme: domains.ThemeDark,
					}).
					Return(nil, errors.New("update failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to update settings", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(gomock.Any()).
					Return(errors.New("update failed"))
			},
			expectedError: errors.New("update failed"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockSettings, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
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

func TestUseCases_SendVerifyEmailMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockrepositories.MockAuthRepository, *mocks.MockLogger, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name:  "successful send verify email",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "user@example.com").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "send verify email to user with special characters",
			email: "user.name+tag@example.co.uk",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "user.name+tag@example.co.uk").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "send verify email to user with cyrillic domain",
			email: "user@почта.рф",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "user@почта.рф").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "auth service returns error - user not found",
			email: "nonexistent@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("user not found")
				mappedErr := errors.New("user not found")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "nonexistent@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user not found"),
		},
		{
			name:  "auth service returns error - invalid email format",
			email: "invalid-email",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("invalid email format")
				mappedErr := errors.New("invalid email format")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "invalid-email").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("invalid email format"),
		},
		{
			name:  "auth service returns error - email already verified",
			email: "verified@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("email already verified")
				mappedErr := errors.New("email already verified")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "verified@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("email already verified"),
		},
		{
			name:  "auth service returns error - rate limit exceeded",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("rate limit exceeded, try again later")
				mappedErr := errors.New("rate limit exceeded, try again later")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "user@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("rate limit exceeded, try again later"),
		},
		{
			name:  "auth service returns error - network timeout",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("network timeout")
				mappedErr := errors.New("network timeout")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "user@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("network timeout"),
		},
		{
			name:  "empty email address",
			email: "",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("email is required")
				mappedErr := errors.New("email is required")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("email is required"),
		},
		{
			name:  "errors mapper transforms error",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("some internal error")
				mappedErr := errors.New("user-friendly error message")

				mockAuth.EXPECT().
					SendVerifyEmailMessage(gomock.Any(), "user@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send verify email message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user-friendly error message"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			err := uc.SendVerifyEmailMessage(ctx, tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_SendForgetPasswordMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockrepositories.MockAuthRepository, *mocks.MockLogger, *mockerrors.MockErrorsMapper)
		expectedError error
	}{
		{
			name:  "successful send forget password message",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "user@example.com").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "send forget password message to user with special characters",
			email: "user.name+tag@example.co.uk",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "user.name+tag@example.co.uk").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "user not found error",
			email: "nonexistent@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("user not found")
				mappedErr := errors.New("user not found")

				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "nonexistent@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send forget password message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user not found"),
		},
		{
			name:  "invalid email format error",
			email: "invalid-email",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("invalid email format")
				mappedErr := errors.New("invalid email format")

				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "invalid-email").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send forget password message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("invalid email format"),
		},
		{
			name:  "rate limit exceeded error",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("rate limit exceeded, try again later")
				mappedErr := errors.New("rate limit exceeded, try again later")

				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "user@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send forget password message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("rate limit exceeded, try again later"),
		},
		{
			name:  "network timeout error",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("network timeout")
				mappedErr := errors.New("network timeout")

				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "user@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send forget password message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("network timeout"),
		},
		{
			name:  "empty email address",
			email: "",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("email is required")
				mappedErr := errors.New("email is required")

				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send forget password message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("email is required"),
		},
		{
			name:  "errors mapper transforms error",
			email: "user@example.com",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("some internal error")
				mappedErr := errors.New("user-friendly error message")

				mockAuth.EXPECT().
					SendForgetPasswordMessage(gomock.Any(), "user@example.com").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to send forget password message",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user-friendly error message"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			err := uc.SendForgetPasswordMessage(ctx, tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_ForgetPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		forgetPasswordToken string
		newPassword         string
		setupMocks          func(*mockrepositories.MockAuthRepository, *mocks.MockLogger, *mockerrors.MockErrorsMapper)
		expectedError       error
	}{
		{
			name:                "successful forget password",
			forgetPasswordToken: "valid-token-123",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token-123", "NewPassword123!").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:                "successful forget password with long password",
			forgetPasswordToken: "valid-token-456",
			newPassword:         "VeryLongPasswordWithManyCharacters123!@#",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token-456", "VeryLongPasswordWithManyCharacters123!@#").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:                "invalid token error",
			forgetPasswordToken: "invalid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("invalid or expired token")
				mappedErr := errors.New("invalid or expired token")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "invalid-token", "NewPassword123!").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("invalid or expired token"),
		},
		{
			name:                "expired token error",
			forgetPasswordToken: "expired-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("token has expired")
				mappedErr := errors.New("token has expired")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "expired-token", "NewPassword123!").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("token has expired"),
		},
		{
			name:                "weak password error",
			forgetPasswordToken: "valid-token",
			newPassword:         "weak",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("password does not meet security requirements")
				mappedErr := errors.New("password does not meet security requirements")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "weak").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("password does not meet security requirements"),
		},
		{
			name:                "user not found error",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("user not found")
				mappedErr := errors.New("user not found")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "NewPassword123!").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user not found"),
		},
		{
			name:                "network error",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("network error")
				mappedErr := errors.New("network error")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "NewPassword123!").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("network error"),
		},
		{
			name:                "empty token",
			forgetPasswordToken: "",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("token is required")
				mappedErr := errors.New("token is required")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "", "NewPassword123!").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("token is required"),
		},
		{
			name:                "empty new password",
			forgetPasswordToken: "valid-token",
			newPassword:         "",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("new password is required")
				mappedErr := errors.New("new password is required")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("new password is required"),
		},
		{
			name:                "errors mapper transforms error",
			forgetPasswordToken: "valid-token",
			newPassword:         "NewPassword123!",
			setupMocks: func(
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("some internal error")
				mappedErr := errors.New("user-friendly error message")

				mockAuth.EXPECT().
					ForgetPassword(gomock.Any(), "valid-token", "NewPassword123!").
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to forget password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user-friendly error message"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			err := uc.ForgetPassword(ctx, tt.forgetPasswordToken, tt.newPassword)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_ChangePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		changePasswordData domains.ChangePasswordDTO
		setupMocks         func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockAuthRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError error
	}{
		{
			name: "successful change password",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					ChangePassword(gomock.Any(), "access_token", domains.ChangePasswordDTO{
						OldPassword: "OldPassword123!",
						NewPassword: "NewPassword123!",
					}).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedError: errors.New("test"),
		},
		{
			name: "failed to change password - wrong password",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "WrongPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("wrong password")
				mappedErr := errors.New("wrong password")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					ChangePassword(gomock.Any(), "access_token", gomock.Any()).
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to change password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("wrong password"),
		},
		{
			name: "network error",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("network error")
				mappedErr := errors.New("network error")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					ChangePassword(gomock.Any(), "access_token", gomock.Any()).
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to change password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("network error"),
		},
		{
			name: "errors mapper transforms error",
			changePasswordData: domains.ChangePasswordDTO{
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword123!",
			},
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockAuth *mockrepositories.MockAuthRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				originalErr := errors.New("some internal error")
				mappedErr := errors.New("user-friendly error message")

				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockAuth.EXPECT().
					ChangePassword(gomock.Any(), "access_token", gomock.Any()).
					Return(originalErr)

				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to change password",
						gomock.Any(),
					).
					Times(1)

				mockErrorsMapper.EXPECT().
					Map(originalErr).
					Return(mappedErr)
			},
			expectedError: errors.New("user-friendly error message"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockAuth, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			ctx := context.Background()
			err := uc.ChangePassword(ctx, tt.changePasswordData)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_UpdateMessage(t *testing.T) {
	t.Parallel()

	dto := domains.UpdateMessageDTO{
		MessageID: 1,
		Text:      "updated text",
	}

	tests := []struct {
		name       string
		dto        domains.UpdateMessageDTO
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError error
	}{
		{
			name: "successful update",
			dto:  dto,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					UpdateMessage(gomock.Any(), "access_token", dto).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			dto:  dto,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to update message",
			dto:  dto,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					UpdateMessage(gomock.Any(), "access_token", dto).
					Return(errors.New("update failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to update message", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedError: errors.New("update failed"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			err := uc.UpdateMessage(
				context.Background(),
				tt.dto,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_DeleteMessage(t *testing.T) {
	t.Parallel()

	dto := domains.DeleteMessageDTO{
		MessageID: 1,
		ForAll:    false,
	}

	tests := []struct {
		name       string
		dto        domains.DeleteMessageDTO
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError error
	}{
		{
			name: "successful delete",
			dto:  dto,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					DeleteMessage(gomock.Any(), "access_token", dto).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "failed to load tokens",
			dto:  dto,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedError: errors.New("tokens not found"),
		},
		{
			name: "failed to delete message",
			dto:  dto,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					DeleteMessage(gomock.Any(), "access_token", dto).
					Return(errors.New("delete failed"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to delete message", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedError: errors.New("delete failed"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			err := uc.DeleteMessage(
				context.Background(),
				tt.dto,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_GetMessageByID(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		messageID  uint64
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedMessage *domains.Message
		expectedError   error
	}{
		{
			name:      "successful get message",
			messageID: 1,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					GetMessageByID(gomock.Any(), "access_token", uint64(1)).
					Return(&domains.Message{
						ID:        1,
						ChatID:    100,
						Text:      "Hello!",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedMessage: &domains.Message{
				ID:        1,
				ChatID:    100,
				Text:      "Hello!",
				CreatedAt: now.In(common.Timezone),
				UpdatedAt: now.In(common.Timezone),
			},
			expectedError: nil,
		},
		{
			name:      "failed to load tokens",
			messageID: 1,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("tokens not found"),
		},
		{
			name:      "failed to get message",
			messageID: 1,
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)

				mockMessages.EXPECT().
					GetMessageByID(gomock.Any(), "access_token", uint64(1)).
					Return(nil, errors.New("not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to get message by id", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("test")).Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("not found"),
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
			mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
			mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			uc := usecases.New(
				mockUsers,
				mockChats,
				mockMessages,
				mockAuth,
				mockTokens,
				mockSettings,
				mockWS,
				mockLogger,
				mockErrorsMapper,
			)

			message, err := uc.GetMessageByID(
				context.Background(),
				tt.messageID,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, message)
			}
		})
	}
}

func newUseCasesForTest(t *testing.T) (
	*usecases.UseCases,
	*mockrepositories.MockTokensRepository,
	*mockrepositories.MockMessagesRepository,
	*mocks.MockLogger,
	*mockerrors.MockErrorsMapper,
) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockTokens := mockrepositories.NewMockTokensRepository(ctrl)
	mockSettings := mockrepositories.NewMockSettingsRepository(ctrl)
	mockUsers := mockrepositories.NewMockUsersRepository(ctrl)
	mockAuth := mockrepositories.NewMockAuthRepository(ctrl)
	mockChats := mockrepositories.NewMockChatsRepository(ctrl)
	mockMessages := mockrepositories.NewMockMessagesRepository(ctrl)
	mockWS := mockrepositories.NewMockWebSocketsRepository(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mockErrorsMapper := mockerrors.NewMockErrorsMapper(ctrl)

	uc := usecases.New(
		mockUsers,
		mockChats,
		mockMessages,
		mockAuth,
		mockTokens,
		mockSettings,
		mockWS,
		mockLogger,
		mockErrorsMapper,
	)

	return uc, mockTokens, mockMessages, mockLogger, mockErrorsMapper
}

func TestUseCases_ListReactions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedReactions []domains.Reaction
		expectedError     bool
	}{
		{
			name: "successful list",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "access_token"}, nil)

				mockMessages.EXPECT().
					ListReactions(gomock.Any(), "access_token").
					Return([]domains.Reaction{{ID: 1, Emoji: "👍"}, {ID: 2, Emoji: "❤️"}}, nil)
			},
			expectedReactions: []domains.Reaction{{ID: 1, Emoji: "👍"}, {ID: 2, Emoji: "❤️"}},
			expectedError:     false,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(nil, errors.New("tokens not found"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("mapped")).Times(1)
			},
			expectedError: true,
		},
		{
			name: "repository error",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "access_token"}, nil)

				mockMessages.EXPECT().
					ListReactions(gomock.Any(), "access_token").
					Return(nil, errors.New("boom"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to list reactions", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("mapped")).Times(1)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, mockTokens, mockMessages, mockLogger, mockErrorsMapper := newUseCasesForTest(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			got, err := uc.ListReactions(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedReactions, got)
			}
		})
	}
}

func TestUseCases_AddMessageReaction(t *testing.T) {
	t.Parallel()

	dto := domains.MessageReactionDTO{MessageID: 10, ReactionID: 1}

	tests := []struct {
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError bool
	}{
		{
			name: "successful add",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "access_token"}, nil)

				mockMessages.EXPECT().
					AddMessageReaction(gomock.Any(), "access_token", dto).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().Load(gomock.Any()).Return(nil, errors.New("tokens"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("mapped")).Times(1)
			},
			expectedError: true,
		},
		{
			name: "repository error (e.g. conflict)",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "access_token"}, nil)

				mockMessages.EXPECT().
					AddMessageReaction(gomock.Any(), "access_token", dto).
					Return(errors.New("already exists"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to add message reaction", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("mapped")).Times(1)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, mockTokens, mockMessages, mockLogger, mockErrorsMapper := newUseCasesForTest(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			err := uc.AddMessageReaction(context.Background(), dto)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUseCases_RemoveMessageReaction(t *testing.T) {
	t.Parallel()

	dto := domains.MessageReactionDTO{MessageID: 10, ReactionID: 1}

	tests := []struct {
		name       string
		setupMocks func(
			*mockrepositories.MockTokensRepository,
			*mockrepositories.MockMessagesRepository,
			*mocks.MockLogger,
			*mockerrors.MockErrorsMapper,
		)
		expectedError bool
	}{
		{
			name: "successful remove",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				_ *mocks.MockLogger,
				_ *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "access_token"}, nil)

				mockMessages.EXPECT().
					RemoveMessageReaction(gomock.Any(), "access_token", dto).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "failed to load tokens",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				_ *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().Load(gomock.Any()).Return(nil, errors.New("tokens"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to load tokens from file", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("mapped")).Times(1)
			},
			expectedError: true,
		},
		{
			name: "repository error",
			setupMocks: func(
				mockTokens *mockrepositories.MockTokensRepository,
				mockMessages *mockrepositories.MockMessagesRepository,
				mockLogger *mocks.MockLogger,
				mockErrorsMapper *mockerrors.MockErrorsMapper,
			) {
				mockTokens.EXPECT().
					Load(gomock.Any()).
					Return(&domains.TokensDTO{AccessToken: "access_token"}, nil)

				mockMessages.EXPECT().
					RemoveMessageReaction(gomock.Any(), "access_token", dto).
					Return(errors.New("boom"))

				mockLogger.EXPECT().
					ErrorContext(gomock.Any(), "failed to remove message reaction", gomock.Any()).
					Times(1)

				mockErrorsMapper.EXPECT().Map(gomock.Any()).Return(errors.New("mapped")).Times(1)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, mockTokens, mockMessages, mockLogger, mockErrorsMapper := newUseCasesForTest(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockTokens, mockMessages, mockLogger, mockErrorsMapper)
			}

			err := uc.RemoveMessageReaction(context.Background(), dto)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
