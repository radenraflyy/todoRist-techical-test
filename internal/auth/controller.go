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

// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Register user payload"
// @Success 201 {object} auth.SuccessResponse "Success response"
// @Failure 422 {object} exception.CustomException "Validation errors"
// @Router /auth/register [post]
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


// LoginUser godoc
// @Summary Login user
// @Description Login with email and password, returns tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "Login user payload"
// @Success 200 {object} LoginResponse "Success login response with tokens"
// @Failure 422 {object} exception.CustomException "Validation errors"
// @Failure 500 {object} auth.ErrorResponse "Internal error"
// @Router /auth/login [post]
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

// LogoutUsers godoc
// @Summary Logout user
// @Description Logout user by clearing refresh token cookie
// @Tags auth
// @Produce json
// @Success 200 {object} auth.SuccessResponse "Success logout response"
// @Router /auth/logout [get]
func (ac *authController) LogoutUsers(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", os.Getenv("GO_ENV") == "production", true)

	utils.SuccessWithoutData(c, http.StatusOK, "success logout!")
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token cookie
// @Tags auth
// @Produce json
// @Success 200 {object} RefreshTokenResponse "New access token response"
// @Failure 401 {object} auth.ErrorResponse "Unauthorized or invalid token"
// @Router /auth/refresh-token [get]
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
