package interfaces

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

//go:generate mockgen -source=usecases.go -destination=../../mocks/usecases/usecases.go -package=mockusecases -exclude_interfaces=
type UseCases interface {
	// Auth
	Authenticate(ctx context.Context) (*domains.User, error)
	GetCurrentUser(ctx context.Context) (*domains.User, error)
	Login(ctx context.Context, email, password string) (*domains.User, error)
	Logout(ctx context.Context) error
	Register(ctx context.Context, registerData domains.RegisterDTO) (*domains.User, error)
	RefreshTokens(ctx context.Context) (*domains.TokensDTO, error)
	SendVerifyEmailMessage(ctx context.Context, email string) error
	SendForgetPasswordMessage(ctx context.Context, email string) error
	ForgetPassword(ctx context.Context, forgetPasswordToken, newPassword string) error

	// Messaging
	SendMessage(ctx context.Context, message domains.Message) error
	ReadMessage(ctx context.Context) (*domains.Message, error)
	GetChatMessages(
		ctx context.Context,
		chatID uint64,
		pagination *domains.Pagination,
	) ([]domains.Message, error)

	// Chats
	CreateChat(ctx context.Context, chat domains.Chat) (*domains.Chat, error)
	GetUserChats(ctx context.Context, pagination *domains.Pagination) ([]domains.Chat, error)

	// Users
	SearchUsers(
		ctx context.Context,
		filters *domains.UsersFilters,
		pagination *domains.Pagination,
	) ([]domains.User, error)

	// Settings
	GetTheme(ctx context.Context) domains.ThemeType
	SetTheme(ctx context.Context, theme domains.ThemeType) error
}
