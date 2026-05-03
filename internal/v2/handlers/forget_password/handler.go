package forget_password

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/errors"
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

func (h *Handler) ForgetPassword(
	forgetPasswordToken string,
	in domains.ForgetPasswordDTO,
) error {
	ctx := context.Background()

	if forgetPasswordToken == "" {
		return h.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(in.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(errors.ErrInvalidPassword)
	}

	return h.useCases.ForgetPassword(ctx, forgetPasswordToken, in.NewPassword)
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
