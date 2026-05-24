package interfaces

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/auth_repository.go -package=mockrepositories -exclude_interfaces=TokensRepository,UsersRepository,ChatsRepository,MessagesRepository,WebSocketsRepository,SettingsRepository
type AuthRepository interface {
	Register(ctx context.Context, registerData domains.RegisterDTO) (*domains.User, error)
	Login(ctx context.Context, in domains.LoginDTO) (*domains.TokensDTO, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*domains.TokensDTO, error)
	SendVerifyEmailMessage(ctx context.Context, email string) error
	SendForgetPasswordMessage(ctx context.Context, email string) error
	ForgetPassword(ctx context.Context, forgetPasswordToken, newPassword string) error
	ChangePassword(
		ctx context.Context,
		accessToken string,
		changePasswordData domains.ChangePasswordDTO,
	) error
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/tokens_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,UsersRepository,ChatsRepository,MessagesRepository,WebSocketsRepository,SettingsRepository
type TokensRepository interface {
	Save(ctx context.Context, tokens domains.TokensDTO) error
	Load(ctx context.Context) (*domains.TokensDTO, error)
	Delete(_ context.Context) error
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/users_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,ChatsRepository,MessagesRepository,WebSocketsRepository,SettingsRepository
type UsersRepository interface {
	GetCurrentUser(ctx context.Context, accessToken string) (*domains.User, error)
	UpdateUser(
		ctx context.Context,
		accessToken string,
		updateUserData domains.UpdateUserDTO,
	) (*domains.User, error)
	SearchUsers(
		ctx context.Context,
		filters *domains.UsersFilters,
		pagination *domains.Pagination,
	) ([]domains.User, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/chats_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,UsersRepository,MessagesRepository,WebSocketsRepository,SettingsRepository
type ChatsRepository interface {
	GetUserChats(
		ctx context.Context,
		accessToken string,
		pagination *domains.Pagination,
	) ([]domains.Chat, error)
	CreateChat(ctx context.Context, accessToken string, chat domains.Chat) (*domains.Chat, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/messages_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,UsersRepository,ChatsRepository,WebSocketsRepository,SettingsRepository
type MessagesRepository interface {
	GetChatMessages(
		ctx context.Context,
		accessToken string,
		chatID uint64,
		pagination *domains.Pagination,
	) ([]domains.Message, error)
	DeleteMessage(
		ctx context.Context,
		accessToken string,
		dto domains.DeleteMessageDTO,
	) error
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/websockets_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,UsersRepository,ChatsRepository,MessagesRepository,SettingsRepository
type WebSocketsRepository interface {
	Connect(ctx context.Context, accessToken string) error
	Close() error
	ReadEvent(ctx context.Context) (*domains.WSEvent, error)
	WriteMessage(ctx context.Context, message domains.Message) error
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/settings_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,UsersRepository,ChatsRepository,MessagesRepository,WebSocketsRepository,TokensRepository
type SettingsRepository interface {
	GetSettings(ctx context.Context, accessToken string) (*domains.Settings, error)
	UpdateSettings(
		ctx context.Context,
		accessToken string,
		settings domains.Settings,
	) (*domains.Settings, error)
}
