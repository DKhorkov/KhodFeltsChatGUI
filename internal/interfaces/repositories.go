package interfaces

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/auth_repository.go -package=mockrepositories -exclude_interfaces=TokensRepository,UsersRepository,ChatsRepository,WebSocketsRepository
type AuthRepository interface {
	Register(ctx context.Context, registerData domains.RegisterDTO) (*domains.User, error)
	Login(ctx context.Context, email, password string) (*domains.TokensDTO, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*domains.TokensDTO, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/tokens_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,UsersRepository,ChatsRepository,WebSocketsRepository
type TokensRepository interface {
	Save(ctx context.Context, tokens domains.TokensDTO) error
	Load(ctx context.Context) (*domains.TokensDTO, error)
	Delete(_ context.Context) error
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/users_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,ChatsRepository,WebSocketsRepository
type UsersRepository interface {
	GetCurrentUser(ctx context.Context, accessToken string) (*domains.User, error)
	SearchUsers(ctx context.Context, username string, limit, offset int) ([]domains.User, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/chats_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,UsersRepository,WebSocketsRepository
type ChatsRepository interface {
	GetUserChats(ctx context.Context, accessToken string, limit, offset int) ([]domains.Chat, error)
	CreateChat(ctx context.Context, accessToken string, chat domains.Chat) (*domains.Chat, error)
	GetChatMessages(
		ctx context.Context,
		accessToken string,
		chatID uint64,
		limit, offset int,
	) ([]domains.Message, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/websockets_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,UsersRepository,ChatsRepository
type WebSocketsRepository interface {
	Connect(ctx context.Context, accessToken string) error
	Close() error
	ReadMessage(ctx context.Context) (*domains.Message, error)
	WriteMessage(ctx context.Context, message domains.Message) error
}
