package auth

import "errors"

var (
	ErrTokenNotMatch = errors.New("refresh token not match")
	ErrTokenExpired  = errors.New("refresh token has expired")
)

type (
	GetUserModel struct {
		UserId       string `json:"userId" db:"user_id"`
		Name         string `json:"name" db:"name"`
		Email        string `json:"email" db:"email"`
		Password     string `json:"-" db:"password"`
		RefreshToken string `json:"-" db:"refresh_token,omitempty"`
	}

	LoginResponse struct {
		AccessToken  string       `db:"access_token" json:"accessToken"`
		RefreshToken string       `db:"refresh_token" json:"-"`
		User         GetUserModel `json:"user"`
	}

	ExistsResultResponse struct {
		Exists bool `db:"exists"`
	}

	GetRefreshTokenResponse struct {
		ID           string `db:"id"`
		RefreshToken string `db:"refresh_token"`
	}

	RefreshTokenResponse struct {
		AccessToken string `db:"access_token" json:"accessToken"`
		User        GetUserModel
	}

	GenerateTokenResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		UserId       string `json:"userId"`
		Name         string `json:"name"`
	}

	SuccessResponse struct {
		Success    bool   `json:"success" example:"true"`
		StatusCode int    `json:"statusCode" example:"200"`
		Message    string `json:"message" example:"success"`
	}

	SuccessDataResponse struct {
		Success    bool        `json:"success" example:"true"`
		StatusCode int         `json:"statusCode" example:"200"`
		Message    string      `json:"message" example:"success"`
		Data       interface{} `json:"data"`
	}

	ErrorResponse struct {
		Success    bool   `json:"success" example:"false"`
		StatusCode int    `json:"statusCode" example:"400"`
		Message    string `json:"message" example:"error message"`
	}
)
