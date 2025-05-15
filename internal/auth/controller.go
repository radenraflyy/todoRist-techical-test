package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"todorist/pkg/exception"
	"todorist/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type AuthController interface {
	RegisterUser(c *gin.Context)
	LoginUser(c *gin.Context)
	LogoutUsers(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type authController struct {
	useCase UseCase
}

func NewAuthController(authRouter *gin.RouterGroup, useCase UseCase) AuthController {
	controller := &authController{
		useCase: useCase,
	}
	authRouter.POST("/register", controller.RegisterUser)
	authRouter.POST("/login", controller.LoginUser)
	authRouter.GET("/refresh-token", controller.RefreshToken)
	authRouter.GET("/logout", controller.LogoutUsers)
	return controller
}

func (ac *authController) RegisterUser(c *gin.Context) {
	var user RegisterRequest

	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Error(c, http.StatusInternalServerError, err)
		return
	}

	user.Email = strings.ToLower(user.Email)
	validationErr := validate.Struct(user)
	if validationErr != nil {
		var errors []string
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors = append(errors, utils.CustomErrorMessage(err))
		}
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", errors),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	err := ac.useCase.RegisterUser(user)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithoutData(c, http.StatusCreated, "success create users")
}

func (ac *authController) LoginUser(c *gin.Context) {
	var user LoginRequest

	if err := c.ShouldBindJSON(&user); err != nil {
		c.Error(err)
		return
	}

	validationErr := validate.Struct(user)
	if validationErr != nil {
		var errors []string
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors = append(errors, utils.CustomErrorMessage(err))
		}
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", errors),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	res, err := ac.useCase.LoginUser(user)
	if err != nil {
		c.Error(err)
		return
	}

	c.SetCookie("refresh_token", res.RefreshToken, 86400, "/", "", os.Getenv("GO_ENV") == "production", true)
	utils.SuccessWithData(c, http.StatusOK, res, "success login!")
}

func (ac *authController) LogoutUsers(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", os.Getenv("GO_ENV") == "production", true)

	utils.SuccessWithoutData(c, http.StatusOK, "success logout!")
}

func (ac *authController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.Error(err)
		return
	}
	log.Println("refresh token:", refreshToken)

	res, err := ac.useCase.RefreshToken(refreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	utils.SuccessWithData(c, http.StatusOK, res, "succesfully refresh token!")
}
