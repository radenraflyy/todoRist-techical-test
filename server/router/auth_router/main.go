package authrouter

import (
	"todorist/config"
	"todorist/internal/auth"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.RouterGroup, db *config.DB) {
	authRouter := r.Group("/auth")

	repository := auth.NewAuthRepository(db)
	useCase := auth.NewUseCase(repository, db)
	auth.NewAuthController(authRouter, useCase)
}
