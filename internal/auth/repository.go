package auth

import (
	"todorist/config"
	"todorist/pkg/customlog"
	hashfunction "todorist/pkg/hash-function"
)

type AuthRepository interface {
	RegisterUser(data RegisterRequest) error
	IsEmailExists(email string) (ExistsResultResponse, error)
	UpdateRefreshToken(data UpdateRefreshTokenRequest) error
	GetRefreshToken(userId string) (GetRefreshTokenResponse, error)
	GetUserByEmail(email string) (GetUserModel, error)
}

type authRepository struct {
	db *config.DB
}

func NewAuthRepository(db *config.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) GetUserByEmail(email string) (GetUserModel, error) {
	var data GetUserModel
	q := `SELECT id AS "user_id", name, email, password FROM "users" WHERE email = $<email>`
	if err := r.db.SelectOne(q, &data, map[string]any{"email": email}); err != nil {
		return data, err
	}
	customlog.PrintJSON(data, "data AING")
	return data, nil
}

func (r *authRepository) RegisterUser(body RegisterRequest) error {
	body.Password = hashfunction.HashPassword(body.Password)
	err := r.db.InsertOne(body, "users", nil, config.WithoutUserId())
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) IsEmailExists(email string) (ExistsResultResponse, error) {
	var result ExistsResultResponse
	query := "SELECT EXISTS (SELECT email FROM users WHERE email = $<email>) as exists"
	err := r.db.SelectOne(query, &result, map[string]any{"email": email})
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *authRepository) UpdateRefreshToken(data UpdateRefreshTokenRequest) error {
	dataUpdated := struct {
		RefreshToken string `db:"refresh_token"`
	}{data.RefreshToken}
	if err := r.db.Update(&dataUpdated, "users", "id = $<id>", map[string]any{"id": data.ID}, nil, config.WithoutUserId()); err != nil {
		return err
	}
	return nil
}

func (r *authRepository) GetRefreshToken(userId string) (GetRefreshTokenResponse, error) {
	var data GetRefreshTokenResponse
	var params map[string]any
	q := `SELECT id, refresh_token FROM "users"
				WHERE id = $<id>`
	params = map[string]any{"id": userId}

	if err := r.db.SelectOne(q, &data, params); err != nil {
		return data, err
	}

	if data.ID == "" {
		return data, nil
	}

	return data, nil
}
