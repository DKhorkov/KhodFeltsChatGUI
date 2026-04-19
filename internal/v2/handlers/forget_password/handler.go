package forget_password

import (
	"context"
	"strconv"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/security"
	"github.com/DKhorkov/libs/validation"
)

type Handler struct {
	useCases         interfaces.UseCases
	errorsMapper     interfaces.ErrorsMapper
	validationConfig config.ValidationConfig
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

type ForgetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) ValidateToken(ctx context.Context, token string) error {
	bytesUserID, err := security.RawDecode(token)
	if err != nil {
		return h.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
	}

	if _, err = strconv.ParseUint(string(bytesUserID), 10, 64); err != nil {
		return h.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
	}

	return nil
}

func (h *Handler) ResetPassword(
	ctx context.Context,
	req ForgetPasswordRequest,
) error {
	// Валидация токена
	if err := h.ValidateToken(ctx, req.Token); err != nil {
		return err
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(req.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(errors.ErrInvalidPassword)
	}

	return h.useCases.ForgetPassword(ctx, req.Token, req.NewPassword)
}
