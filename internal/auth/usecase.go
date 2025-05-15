package auth

import (
	"net/http"
	"time"
	"todorist/config"
	"todorist/pkg/customlog"
	"todorist/pkg/exception"
	"todorist/pkg/jwttoken"
	verifypassword "todorist/pkg/verify-password"
)

const (
	expireAccessToken  = 10 * time.Minute
	expireRefreshToken = 24 * time.Hour
)

type UseCase interface {
	RegisterUser(users RegisterRequest) error
	LoginUser(data LoginRequest) (LoginResponse, error)
	RefreshToken(refreshToken string) (*RefreshTokenResponse, error)
	GenerateToken(user GetUserModel) (*GenerateTokenResponse, error)
}

type useCase struct {
	repo AuthRepository
	db   *config.DB
}

func NewUseCase(repo AuthRepository, db *config.DB) UseCase {
	return &useCase{
		repo: repo,
		db:   db,
	}
}

func (us *useCase) RegisterUser(data RegisterRequest) error {
	result, err := us.repo.IsEmailExists(data.Email)
	if err != nil {
		return err
	}
	if result.Exists {
		return &exception.BadRequestException{
			Message: "Email telah terdaftar. Silakan gunakan email lain",
		}
	}

	err = us.repo.RegisterUser(data)
	return err
}

func (us *useCase) LoginUser(data LoginRequest) (LoginResponse, error) {
	dataUser, err := us.repo.GetUserByEmail(data.Email)
	if err != nil {
		return LoginResponse{}, &exception.NotFoundException{
			Message: err.Error(),
		}
	}

	err = verifypassword.VerifyPassword(data.Password, dataUser.Password)
	if err != nil {
		return LoginResponse{}, &exception.BadRequestException{
			Message: "Password Salah",
		}
	}

	token, err := us.GenerateToken(dataUser)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		User:         dataUser,
	}, nil
}

func (us *useCase) GenerateToken(user GetUserModel) (*GenerateTokenResponse, error) {
	claims := map[string]interface{}{
		"user_id":         user.UserId,
		"name":            user.Name,
		"email":           user.Email,
		"profile_picture": user.ProfilePicture,
	}

	accesToken, err := jwttoken.CreateToken(expireAccessToken, claims)
	if err != nil {
		return nil, &exception.CustomException{
			Message: err.Error(),
			Code:    http.StatusNotFound,
		}
	}

	refreshToken, err := jwttoken.CreateToken(expireRefreshToken, claims)
	if err != nil {
		return nil, &exception.CustomException{
			Message: err.Error(),
			Code:    http.StatusNotFound,
		}
	}

	dataUpSertRefreshToken := UpdateRefreshTokenRequest{
		ID:           user.UserId,
		RefreshToken: refreshToken,
	}
	err = us.repo.UpdateRefreshToken(dataUpSertRefreshToken)
	if err != nil {
		return nil, &exception.CustomException{
			Message: "Error update refreshToken in DB",
			Code:    http.StatusBadRequest,
		}
	}

	return &GenerateTokenResponse{
		AccessToken:  accesToken,
		RefreshToken: refreshToken,
	}, nil
}

func (us *useCase) RefreshToken(refreshToken string) (*RefreshTokenResponse, error) {
	user, err := jwttoken.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrTokenExpired
	}
	customlog.PrintJSON(user, "user AING")
	existingRefreshToken, err := us.repo.GetRefreshToken(user.UserId)
	if err != nil {
		return nil, err
	}

	if existingRefreshToken.RefreshToken == "" {
		return nil, &exception.CustomException{
			Message: "TOKEN_EXPIRED",
			Code:    http.StatusUnauthorized,
		}
	}

	if existingRefreshToken.RefreshToken != refreshToken {
		return nil, &exception.CustomException{
			Message: "INVALID_TOKEN",
			Code:    http.StatusUnauthorized,
		}
	}

	accessToken, err := jwttoken.CreateToken(expireAccessToken, map[string]interface{}{
		"name":            user.Name,
		"email":           user.Email,
		"profile_picture": user.ProfilePicuture,
		"user_id":         user.UserId,
	})

	if err != nil {
		return nil, err
	}

	response := &RefreshTokenResponse{
		AccessToken: accessToken,
		User: GetUserModel{
			UserId: user.UserId,
			Name:   user.Name,
			Email:  user.Email,
		},
	}

	return response, nil
}
