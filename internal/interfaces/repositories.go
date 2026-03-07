package interfaces

import (
	"context"
	"kfcGUI/internal/domains"
)

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/auth_repository.go -package=mockrepositories -exclude_interfaces=UsersRepository,MessagesRepository,ChatsRepository,WebSocketsRepository
type AuthRepository interface {
	Register(ctx context.Context, user domains.User) (*domains.User, error)
	Login(ctx context.Context, email, password string) (*domains.TokensDTO, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*domains.TokensDTO, error)
}
