package usecases

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
)

type UseCases struct {
	users        interfaces.UsersRepository
	chats        interfaces.ChatsRepository
	messages     interfaces.MessagesRepository
	auth         interfaces.AuthRepository
	tokens       interfaces.TokensRepository
	settings     interfaces.SettingsRepository
	ws           interfaces.WebSocketsRepository
	logger       logging.Logger
	errorsMapper interfaces.ErrorsMapper
}

func New(
	users interfaces.UsersRepository,
	chats interfaces.ChatsRepository,
	messages interfaces.MessagesRepository,
	auth interfaces.AuthRepository,
	tokens interfaces.TokensRepository,
	settings interfaces.SettingsRepository,
	ws interfaces.WebSocketsRepository,
	logger logging.Logger,
	errorsMapper interfaces.ErrorsMapper,
) *UseCases {
	return &UseCases{
		users:        users,
		chats:        chats,
		messages:     messages,
		auth:         auth,
		tokens:       tokens,
		settings:     settings,
		ws:           ws,
		logger:       logger,
		errorsMapper: errorsMapper,
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

		return nil, u.errorsMapper.Map(err)
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

		return nil, u.errorsMapper.Map(err)
	}

	return user, nil
}

func (u *UseCases) RefreshTokens(ctx context.Context) (*domains.TokensDTO, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	tokens, err = u.auth.RefreshTokens(ctx, tokens.RefreshToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to refresh tokens", err)

		return nil, u.errorsMapper.Map(err)
	}

	if err = u.tokens.Save(ctx, *tokens); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to save tokens", err)

		return nil, u.errorsMapper.Map(err)
	}

	return tokens, nil
}

func (u *UseCases) Login(ctx context.Context, loginData domains.LoginDTO) (*domains.User, error) {
	tokens, err := u.auth.Login(ctx, loginData)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to login", err)

		return nil, u.errorsMapper.Map(err)
	}

	if err = u.tokens.Save(ctx, *tokens); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to save tokens", err)

		return nil, u.errorsMapper.Map(err)
	}

	user, err := u.users.GetCurrentUser(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get user by accessToken", err)

		return nil, u.errorsMapper.Map(err)
	}

	return user, nil
}

func (u *UseCases) Logout(ctx context.Context) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.auth.Logout(ctx, tokens.AccessToken); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to logout", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.tokens.Delete(ctx); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to delete tokens", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.ws.Close(); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to close websockets", err)

		return u.errorsMapper.Map(err)
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

		return nil, u.errorsMapper.Map(err)
	}

	return user, nil
}

func (u *UseCases) SendVerifyEmailMessage(ctx context.Context, email string) error {
	if err := u.auth.SendVerifyEmailMessage(ctx, email); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to send verify email message", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) SendForgetPasswordMessage(ctx context.Context, email string) error {
	if err := u.auth.SendForgetPasswordMessage(ctx, email); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to send forget password message", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) ChangePassword(
	ctx context.Context,
	changePasswordData domains.ChangePasswordDTO,
) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.auth.ChangePassword(ctx, tokens.AccessToken, changePasswordData); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to change password", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) ForgetPassword(
	ctx context.Context,
	forgetPasswordToken, newPassword string,
) error {
	if err := u.auth.ForgetPassword(ctx, forgetPasswordToken, newPassword); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to forget password", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) SendMessage(ctx context.Context, message domains.Message) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.ws.Connect(ctx, tokens.AccessToken); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to connect to websockets", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.ws.WriteMessage(ctx, message); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to send message", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) ReadEvent(ctx context.Context) (*domains.WSEvent, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	if err = u.ws.Connect(ctx, tokens.AccessToken); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to connect to websockets", err)

		return nil, u.errorsMapper.Map(err)
	}

	event, err := u.ws.ReadEvent(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to read event", err)

		return nil, u.errorsMapper.Map(err)
	}

	return event, nil
}

func (u *UseCases) GetMessageByID(ctx context.Context, messageID uint64) (*domains.Message, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	message, err := u.messages.GetMessageByID(ctx, tokens.AccessToken, messageID)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get message by id", err)

		return nil, u.errorsMapper.Map(err)
	}

	message.CreatedAt = message.CreatedAt.In(common.Timezone)
	message.UpdatedAt = message.UpdatedAt.In(common.Timezone)

	return message, nil
}

func (u *UseCases) DeleteMessage(ctx context.Context, dto domains.DeleteMessageDTO) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return u.errorsMapper.Map(err)
	}

	if err = u.messages.DeleteMessage(ctx, tokens.AccessToken, dto); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to delete message", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) CreateChat(ctx context.Context, chat domains.Chat) (*domains.Chat, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	createdChat, err := u.chats.CreateChat(ctx, tokens.AccessToken, chat)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to create chat", err)

		return nil, u.errorsMapper.Map(err)
	}

	return createdChat, nil
}

func (u *UseCases) GetUserChats(
	ctx context.Context,
	pagination *domains.Pagination,
) ([]domains.Chat, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	chats, err := u.chats.GetUserChats(ctx, tokens.AccessToken, pagination)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get user chats", err)

		return nil, u.errorsMapper.Map(err)
	}

	return chats, nil
}

func (u *UseCases) UpdateUser(
	ctx context.Context,
	updateUserData domains.UpdateUserDTO,
) (*domains.User, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	updatedUser, err := u.users.UpdateUser(ctx, tokens.AccessToken, updateUserData)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to update user", err)

		return nil, u.errorsMapper.Map(err)
	}

	return updatedUser, nil
}

func (u *UseCases) SearchUsers(
	ctx context.Context,
	filters *domains.UsersFilters,
	pagination *domains.Pagination,
) ([]domains.User, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	currentUser, err := u.users.GetCurrentUser(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get current user", err)

		return nil, u.errorsMapper.Map(err)
	}

	users, err := u.users.SearchUsers(ctx, filters, pagination)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to search users", err)

		return nil, u.errorsMapper.Map(err)
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
	pagination *domains.Pagination,
) ([]domains.Message, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	messages, err := u.messages.GetChatMessages(ctx, tokens.AccessToken, chatID, pagination)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get chat messages", err)

		return nil, u.errorsMapper.Map(err)
	}

	// Устанавливаем время сообщений по местному времени польователя
	for i := range messages {
		messages[i].CreatedAt = messages[i].CreatedAt.In(common.Timezone)
		messages[i].UpdatedAt = messages[i].UpdatedAt.In(common.Timezone)
	}

	return messages, nil
}

func (u *UseCases) GetTheme(ctx context.Context) domains.ThemeType {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return domains.ThemeLight // default theme
	}

	settings, err := u.settings.GetSettings(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get settings", err)

		return domains.ThemeLight // default theme
	}

	return settings.Theme
}

func (u *UseCases) SetTheme(ctx context.Context, theme domains.ThemeType) error {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return u.errorsMapper.Map(err)
	}

	currentSettings, err := u.settings.GetSettings(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get settings", err)

		return u.errorsMapper.Map(err)
	}

	currentSettings.Theme = theme

	if _, err = u.settings.UpdateSettings(ctx, tokens.AccessToken, *currentSettings); err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to update settings", err)

		return u.errorsMapper.Map(err)
	}

	return nil
}

func (u *UseCases) GetSettings(ctx context.Context) (*domains.Settings, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	settings, err := u.settings.GetSettings(ctx, tokens.AccessToken)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to get settings", err)

		return nil, u.errorsMapper.Map(err)
	}

	return settings, nil
}

func (u *UseCases) UpdateSettings(
	ctx context.Context,
	settings domains.Settings,
) (*domains.Settings, error) {
	tokens, err := u.tokens.Load(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to load tokens from file", err)

		return nil, u.errorsMapper.Map(err)
	}

	updatedSettings, err := u.settings.UpdateSettings(ctx, tokens.AccessToken, settings)
	if err != nil {
		logging.LogErrorContext(ctx, u.logger, "failed to update settings", err)

		return nil, u.errorsMapper.Map(err)
	}

	return updatedSettings, nil
}
