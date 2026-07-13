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
	Login(ctx context.Context, in domains.LoginDTO) (*domains.User, error)
	Logout(ctx context.Context) error
	Register(ctx context.Context, registerData domains.RegisterDTO) (*domains.User, error)
	RefreshTokens(ctx context.Context) (*domains.TokensDTO, error)
	SendVerifyEmailMessage(ctx context.Context, email string) error
	SendForgetPasswordMessage(ctx context.Context, email string) error
	ForgetPassword(ctx context.Context, forgetPasswordToken, newPassword string) error
	ChangePassword(ctx context.Context, changePasswordData domains.ChangePasswordDTO) error

	// Messaging
	SendMessage(ctx context.Context, message domains.Message) error
	ReadEvent(ctx context.Context) (*domains.WSEvent, error)
	UpdateMessage(ctx context.Context, dto domains.UpdateMessageDTO) error
	DeleteMessage(ctx context.Context, dto domains.DeleteMessageDTO) error
	GetMessageByID(ctx context.Context, messageID uint64) (*domains.Message, error)
	GetChatMessages(
		ctx context.Context,
		chatID uint64,
		pagination *domains.Pagination,
	) ([]domains.Message, error)

	// Reactions
	ListReactions(ctx context.Context) ([]domains.Reaction, error)
	AddMessageReaction(ctx context.Context, dto domains.MessageReactionDTO) error
	RemoveMessageReaction(ctx context.Context, dto domains.MessageReactionDTO) error

	// Chats
	CreateChat(ctx context.Context, chat domains.Chat) (*domains.Chat, error)
	GetUserChats(ctx context.Context, pagination *domains.Pagination) ([]domains.Chat, error)

	// Users
	UpdateUser(ctx context.Context, updateUserData domains.UpdateUserDTO) (*domains.User, error)
	SearchUsers(
		ctx context.Context,
		filters *domains.UsersFilters,
		pagination *domains.Pagination,
	) ([]domains.User, error)
	UpdateAvatar(ctx context.Context, fileData []byte) (string, error)
	DeleteAvatar(ctx context.Context) error

	// Settings
	GetTheme(ctx context.Context) domains.ThemeType
	SetTheme(ctx context.Context, theme domains.ThemeType) error
	GetSettings(ctx context.Context) (*domains.Settings, error)
	UpdateSettings(ctx context.Context, settings domains.Settings) (*domains.Settings, error)
}
