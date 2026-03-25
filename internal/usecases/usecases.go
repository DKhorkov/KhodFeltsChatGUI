package usecases

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
)

type UseCases struct {
	users  interfaces.UsersRepository
	chats  interfaces.ChatsRepository
	auth   interfaces.AuthRepository
	tokens interfaces.TokensRepository
	ws     interfaces.WebSocketsRepository
	logger logging.Logger
}

func New(
	users interfaces.UsersRepository,
	chats interfaces.ChatsRepository,
	auth interfaces.AuthRepository,
	tokens interfaces.TokensRepository,
	ws interfaces.WebSocketsRepository,
	logger logging.Logger,
) *UseCases {
	return &UseCases{
		users:  users,
		chats:  chats,
		auth:   auth,
		tokens: tokens,
		ws:     ws,
		logger: logger,
	}
}

func (u *UseCases) Authenticate(ctx context.Context) (*domains.User, error) {
	tokens, err := u.RefreshTokens(ctx)
	if err != nil {
		return nil, err
	}

	user, err := u.users.GetCurrentUser(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get user by refreshed tokens", err)

		return nil, err
	}

	return user, nil
}

func (u *UseCases) GetCurrentUser(ctx context.Context) (*domains.User, error) {
	tokens, err := u.RefreshTokens(ctx)
	if err != nil {
		return nil, err
	}

	user, err := u.users.GetCurrentUser(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get user by refreshed tokens", err)

		return nil, err
	}

	return user, nil
}

func (u *UseCases) RefreshTokens(ctx context.Context) (*domains.TokensDTO, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, err
	}

	tokens, err = u.auth.RefreshTokens(ctx, tokens.RefreshToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to refresh tokens", err)

		return nil, err
	}

	if err = u.tokens.Save(ctx, *tokens); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to save tokens", err)

		return nil, err
	}

	return tokens, nil
}

func (u *UseCases) Login(ctx context.Context, email, password string) (*domains.User, error) {
	tokens, err := u.auth.Login(ctx, email, password)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to login", err)

		return nil, err
	}

	if err = u.tokens.Save(ctx, *tokens); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to save tokens", err)

		return nil, err
	}

	user, err := u.users.GetCurrentUser(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get user by accessToken", err)

		return nil, err
	}

	return user, nil
}

func (u *UseCases) Logout(ctx context.Context) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return err
	}

	if err = u.auth.Logout(ctx, tokens.AccessToken); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to logout", err)

		return err
	}

	if err = u.tokens.Delete(ctx); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to delete tokens", err)

		return err
	}

	if err = u.ws.Close(); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to close websockets", err)

		return err
	}

	return nil
}

func (u *UseCases) Register(
	ctx context.Context,
	registerData domains.RegisterDTO,
) (*domains.User, error) {
	user, err := u.auth.Register(ctx, registerData)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to register user", err)

		return nil, err
	}

	return user, nil
}

func (u *UseCases) SendMessage(ctx context.Context, message domains.Message) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return err
	}

	if err = u.ws.Connect(ctx, tokens.AccessToken); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to connect to websockets", err)

		return err
	}

	if err = u.ws.WriteMessage(ctx, message); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to send message", err)

		return err
	}

	return nil
}

func (u *UseCases) ReadMessage(ctx context.Context) (*domains.Message, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, err
	}

	if err = u.ws.Connect(ctx, tokens.AccessToken); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to connect to websockets", err)

		return nil, err
	}

	message, err := u.ws.ReadMessage(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to read message", err)

		return nil, err
	}

	return message, nil
}

func (u *UseCases) CreateChat(ctx context.Context, chat domains.Chat) (*domains.Chat, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, err
	}

	createdChat, err := u.chats.CreateChat(ctx, tokens.AccessToken, chat)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to create chat", err)

		return nil, err
	}

	return createdChat, nil
}

func (u *UseCases) GetUserChats(ctx context.Context, limit, offset int) ([]domains.Chat, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, err
	}

	chats, err := u.chats.GetUserChats(ctx, tokens.AccessToken, limit, offset)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get user chats", err)

		return nil, err
	}

	return chats, nil
}

func (u *UseCases) SearchUsers(
	ctx context.Context,
	username string,
	limit, offset int,
) ([]domains.User, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, err
	}

	currentUser, err := u.users.GetCurrentUser(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get current user", err)

		return nil, err
	}

	users, err := u.users.SearchUsers(ctx, username, limit, offset)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to search users", err)

		return nil, err
	}

	if len(users) == 0 {
		return users, nil
	}

	// Возвращаем всех юзеров кроме текущего
	otherUsers := make([]domains.User, 0, len(users)-1)

	for _, user := range users {
		if user.ID != currentUser.ID {
			otherUsers = append(otherUsers, user)
		}
	}

	return otherUsers, nil
}

func (u *UseCases) GetChatMessages(
	ctx context.Context,
	chatID uint64,
	limit, offset int,
) ([]domains.Message, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, err
	}

	messages, err := u.chats.GetChatMessages(ctx, tokens.AccessToken, chatID, limit, offset)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get chat messages", err)

		return nil, err
	}

	return messages, nil
}
