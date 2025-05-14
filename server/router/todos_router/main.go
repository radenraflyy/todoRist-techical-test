package todosrouter

import (
	"todorist/config"
	"todorist/internal/todos"
	"todorist/server/middleware"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.RouterGroup, db *config.DB) {
	authRouter := r.Group("/todo")
	authRouter.Use(middleware.AuthMiddleware(db))

	repository := todos.NewTodosRepository(db)
	useCase := todos.NewUseCase(repository, db)
	todos.NewTodosController(authRouter, useCase)
}
