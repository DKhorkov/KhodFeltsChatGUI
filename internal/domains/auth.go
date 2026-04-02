package domains

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokensDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ForgetPasswordDTO struct {
	NewPassword string `json:"newPassword"`
}

type ChangePasswordDTO struct {
	UserID      uint64 `json:"userId"`
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
}

type SendVerifyEmailMessageMessageDTO struct {
	Email string `json:"email"`
}

type SendForgetPasswordMessageDTO struct {
	Email string `json:"email"`
}
