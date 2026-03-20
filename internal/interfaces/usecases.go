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

	// Messaging
	SendMessage(ctx context.Context, message domains.Message) error
	ReadMessage(ctx context.Context) (*domains.Message, error)
	GetChatMessages(ctx context.Context, chatID uint64, limit, offset int) ([]domains.Message, error)

	// Chats
	CreateChat(ctx context.Context, chat domains.Chat) (*domains.Chat, error)
	GetUserChats(ctx context.Context, limit, offset int) ([]domains.Chat, error)

	// Users
	SearchUsers(ctx context.Context, username string, limit, offset int) ([]domains.User, error)
}
