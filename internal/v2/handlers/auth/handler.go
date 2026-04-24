package auth

import (
	"context"
	"errors"

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

	ctx context.Context
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
		ctx:              context.Background(),
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *Handler) Login(email, password string) error {
	// Валидация email
	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(password, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	// Вызов бизнес-логики

	if _, err := h.useCases.Login(h.ctx, email, password); err != nil {
		if errors.Is(err, customerrors.ErrDefault) {
			return h.errorsMapper.Map(customerrors.ErrLogin)
		}

		return err
	}

	return nil
}

func (h *Handler) Register(in domains.RegisterDTO) error {
	// Валидация email
	if !validation.ValidateValueByRule(in.Email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	// Валидация username
	if !validation.ValidateValueByRules(in.Username, h.validationConfig.UsernameRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidUsername)
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(in.Password, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	if _, err := h.useCases.Register(h.ctx, in); err != nil {
		if errors.Is(err, customerrors.ErrDefault) {
			return h.errorsMapper.Map(customerrors.ErrRegister)
		}

		return err
	}

	return nil
}

func (h *Handler) SendVerifyEmail(email string) error {
	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	return h.useCases.SendVerifyEmailMessage(h.ctx, email)
}

func (h *Handler) SendForgetPassword(email string) error {
	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	return h.useCases.SendForgetPasswordMessage(h.ctx, email)
}

func (h *Handler) ForgetPassword(token string, in domains.ForgetPasswordDTO) error {
	// Валидация пароля
	if !validation.ValidateValueByRules(in.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	return h.useCases.ForgetPassword(h.ctx, token, in.NewPassword)
}

func (h *Handler) Authenticate() error {
	if _, err := h.useCases.Authenticate(h.ctx); err != nil {
		return err
	}

	return nil
}

func (h *Handler) Logout() error {
	return h.useCases.Logout(h.ctx)
}
