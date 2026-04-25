package forget_password

import (
	"context"
	"strconv"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/security"
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

	bytesUserID, err := security.RawDecode(forgetPasswordToken)
	if err != nil {
		return h.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
	}

	if _, err = strconv.ParseUint(string(bytesUserID), 10, 64); err != nil {
		return h.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(in.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(errors.ErrInvalidPassword)
	}

	return h.useCases.ForgetPassword(ctx, forgetPasswordToken, in.NewPassword)
}

func (h *Handler) StartListening() {}

func (h *Handler) StopListening() {}
