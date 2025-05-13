package jwttoken

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserId          string `json:"user_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	ProfilePicuture string `json:"profile_picture"`
	Expired         *int64 `json:"exp"`
	jwt.RegisteredClaims
}

type UserTokenInterface struct {
	UserId   string
	Name     string
	Email    string
	Duration time.Duration
}

func NewUserClaims(data UserTokenInterface) (*UserClaims, error) {
	return &UserClaims{
		UserId: data.UserId,
		Name:   data.Name,
		Email:  data.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(data.Duration)),
		},
	}, nil
}
