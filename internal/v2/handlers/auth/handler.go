package auth

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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func (h *Handler) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Валидация email
	if !validation.ValidateValueByRule(req.Email, h.validationConfig.EmailRegExp) {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrInvalidEmail).Error(),
		}, nil
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(req.Password, h.validationConfig.PasswordRegExps) {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrInvalidPassword).Error(),
		}, nil
	}

	// Вызов бизнес-логики
	_, err := h.useCases.Login(ctx, req.Email, req.Password)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrLogin).Error(),
		}, nil
	}

	return &LoginResponse{
		Success: true,
		Message: "Успешный вход",
	}, nil
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(ctx context.Context, req RegisterRequest) (*LoginResponse, error) {
	// Валидация email
	if !validation.ValidateValueByRule(req.Email, h.validationConfig.EmailRegExp) {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrInvalidEmail).Error(),
		}, nil
	}

	// Валидация username
	if !validation.ValidateValueByRules(req.Username, h.validationConfig.UsernameRegExps) {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrInvalidUsername).Error(),
		}, nil
	}

	// Валидация пароля
	if !validation.ValidateValueByRules(req.Password, h.validationConfig.PasswordRegExps) {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrInvalidPassword).Error(),
		}, nil
	}

	// Регистрация
	registerData := domains.RegisterDTO{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	_, err := h.useCases.Register(ctx, registerData)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: h.errorsMapper.Map(errors.ErrRegister).Error(),
		}, nil
	}

	return &LoginResponse{
		Success: true,
		Message: "Регистрация успешна! Теперь войдите.",
	}, nil
}

func (h *Handler) SendVerifyEmail(ctx context.Context, email string) error {
	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(errors.ErrInvalidEmail)
	}

	return h.useCases.SendVerifyEmailMessage(ctx, email)
}

func (h *Handler) SendForgetPassword(ctx context.Context, email string) error {
	if !validation.ValidateValueByRule(email, h.validationConfig.EmailRegExp) {
		return h.errorsMapper.Map(errors.ErrInvalidEmail)
	}

	return h.useCases.SendForgetPasswordMessage(ctx, email)
}

type ForgetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) ForgetPassword(ctx context.Context, req ForgetPasswordRequest) error {
	// Валидация пароля
	if !validation.ValidateValueByRules(req.NewPassword, h.validationConfig.PasswordRegExps) {
		return h.errorsMapper.Map(errors.ErrInvalidPassword)
	}

	return h.useCases.ForgetPassword(ctx, req.Token, req.NewPassword)
}
