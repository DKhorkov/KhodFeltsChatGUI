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

func (h *Handler) Login(email, password string) error {
	ctx := context.Background()

	// Валидация email
	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(password, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	// Вызов бизнес-логики

	if _, err := h.useCases.Login(ctx, email, password); err != nil {
		if errors.Is(err, customerrors.ErrDefault) {
			return h.errorsMapper.Map(customerrors.ErrLogin)
		}

		return err
	}

	return nil
}

func (h *Handler) Register(in domains.RegisterDTO) error {
	ctx := context.Background()

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

	if _, err := h.useCases.Register(ctx, in); err != nil {
		if errors.Is(err, customerrors.ErrDefault) {
			return h.errorsMapper.Map(customerrors.ErrRegister)
		}

		return err
	}

	return nil
}

func (h *Handler) SendVerifyEmail(email string) error {
	ctx := context.Background()

	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	return h.useCases.SendVerifyEmailMessage(ctx, email)
}

func (h *Handler) SendForgetPassword(email string) error {
	ctx := context.Background()

	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(customerrors.ErrInvalidEmail)
	}

	return h.useCases.SendForgetPasswordMessage(ctx, email)
}

func (h *Handler) ForgetPassword(token string, in domains.ForgetPasswordDTO) error {
	ctx := context.Background()

	// Валидация пароля
	if !validation.ValidateValueByRules(in.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(customerrors.ErrInvalidPassword)
	}

	return h.useCases.ForgetPassword(ctx, token, in.NewPassword)
}

func (h *Handler) Authenticate() error {
	ctx := context.Background()

	if _, err := h.useCases.Authenticate(ctx); err != nil {
		return err
	}

	return nil
}

func (h *Handler) Logout() error {
	ctx := context.Background()

	return h.useCases.Logout(ctx)
}

func (h *Handler) StartListening() {}

func (h *Handler) StopListening() {}
