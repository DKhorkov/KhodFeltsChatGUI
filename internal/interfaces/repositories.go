package interfaces

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
)

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/auth_repository.go -package=mockrepositories -exclude_interfaces=TokensRepository,UsersRepository,MessagesRepository,ChatsRepository,WebSocketsRepository
type AuthRepository interface {
	Register(ctx context.Context, user domains.User) (*domains.User, error)
	Login(ctx context.Context, email, password string) (*domains.TokensDTO, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*domains.TokensDTO, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/tokens_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,UsersRepository,MessagesRepository,ChatsRepository,WebSocketsRepository
type TokensRepository interface {
	Save(ctx context.Context, tokens *domains.TokensDTO) error
	Load(ctx context.Context) (*domains.TokensDTO, error)
}

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/users_repository.go -package=mockrepositories -exclude_interfaces=AuthRepository,TokensRepository,MessagesRepository,ChatsRepository,WebSocketsRepository
type UsersRepository interface {
	GetCurrentUser(ctx context.Context, accessToken string) (*domains.User, error)
	SearchUsers(ctx context.Context, username string, limit, offset int) ([]domains.User, error)
}
