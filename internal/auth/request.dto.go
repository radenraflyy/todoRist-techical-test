package auth

type (
	RegisterRequest struct {
		Name     string `json:"name" db:"name" validate:"required,min=6"`
		Email    string `json:"email" db:"email" validate:"required,email"`
		Password string `json:"password" db:"password" validate:"required,min=8"`
	}

	LoginRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" db:"password" validate:"required"`
	}

	UpdateRefreshTokenRequest struct {
		ID           string `json:"userId" db:"id"`
		RefreshToken string `json:"RefreshToken" db:"refresh_token"`
	}
)
