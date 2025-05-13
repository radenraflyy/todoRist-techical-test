package jwttoken

import (
	"strings"
	"time"
	"todorist/env"
	"todorist/pkg/exception"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(expiration time.Duration, customClaims map[string]interface{}) (string, error) {
	var secretKey = []byte(env.JwtScretKey)

	if customClaims == nil {
		customClaims = make(map[string]interface{})
	}

	if expiration > 0 {
		customClaims["exp"] = time.Now().Add(expiration).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(customClaims))

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (dataClaims *UserClaims, err error) {
	var secretKey = []byte(env.JwtScretKey)
	claims := &UserClaims{}
	tokenString := strings.TrimPrefix(signedToken, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		if err.Error() == "token has invalid claims: token is expired" {
			return nil, &exception.UnautorizedException{
				Message: "NEED_REFRESH_TOKEN",
			}
		} else {
			return nil, &exception.UnautorizedException{
				Message: "Invalid token",
			}
		}
	}

	if claims.Expired != nil {
		if time.Now().Unix() > *claims.Expired {
			return nil, &exception.UnautorizedException{
				Message: "Token Expired",
			}
		}
	}

	return claims, nil
}
