package profile

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

func (h *Handler) ChangePassword(in domains.ChangePasswordDTO) error {
	ctx := context.Background()

	// Валидация нового пароля
	if !validation.ValidateValueByRules(in.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	// Валидация старого пароля
	if !validation.ValidateValueByRules(in.OldPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	if err := h.useCases.ChangePassword(ctx, in); err != nil {
		return err
	}

	return nil
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
