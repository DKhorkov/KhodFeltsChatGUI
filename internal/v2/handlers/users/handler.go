package users

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/validation"
)

type Handler struct {
	useCases         interfaces.UseCases
	errorsMapper     interfaces.ErrorsMapper
	validationConfig config.ValidationConfig

	wailsCtx context.Context
}

func New(
	useCases interfaces.UseCases,
	errorsMapper interfaces.ErrorsMapper,
	validationConfig config.ValidationConfig,
) *Handler {
	return &Handler{
		useCases:         useCases,
		errorsMapper:     errorsMapper,
		validationConfig: validationConfig,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) GetCurrentUser() (*domains.User, error) {
	ctx := context.Background()

	return h.useCases.GetCurrentUser(ctx)
}

func (h *Handler) SearchUsers(
	filters *domains.UsersFilters,
	pagination *domains.Pagination,
) ([]domains.User, error) {
	ctx := context.Background()

	return h.useCases.SearchUsers(ctx, filters, pagination)
}

func (h *Handler) UpdateUser(in domains.UpdateUserDTO) (*domains.User, error) {
	ctx := context.Background()

	// Валидация username
	if in.Username != nil &&
		!validation.ValidateValueByRules(*in.Username, h.validationConfig.UsernameRegExps) {
		return nil, h.errorsMapper.Map(customerrors.ErrInvalidLogin)
	}

	updatedUser, err := h.useCases.UpdateUser(ctx, in)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (h *Handler) UpdateAvatar(fileData []byte) (string, error) {
	ctx := context.Background()

	avatarURL, err := h.useCases.UpdateAvatar(ctx, fileData)
	if err != nil {
		return "", err
	}

	return avatarURL, nil
}

func (h *Handler) DeleteAvatar() error {
	ctx := context.Background()

	return h.useCases.DeleteAvatar(ctx)
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
