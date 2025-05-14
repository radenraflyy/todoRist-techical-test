package router

import (
	"net/http"
	"todorist/config"
	"todorist/server/middleware"
	authrouter "todorist/server/router/auth_router"
	todosrouter "todorist/server/router/todos_router"

	"github.com/gin-gonic/gin"
)

type SetupRoutesConfig struct {
	Router *gin.Engine
	DB     *config.DB
}

func SetupRoutes(c SetupRoutesConfig) {
	c.Router.Use(middleware.ErrorHandler())
	c.Router.Use(middleware.CORSMiddleware())

	apiV1 := c.Router.Group("/v1")

	apiV1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Server healthy..."})
	})

	apiV1.GET("/error", func(c *gin.Context) {
		panic("error test panic")
	})

	authrouter.Init(apiV1, c.DB)
	todosrouter.Init(apiV1, c.DB)
}
