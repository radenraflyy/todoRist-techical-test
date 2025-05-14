package todos

import (
	"fmt"
	"net/http"
	"todorist/pkg/exception"
	"todorist/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type TodosController interface {
	CreateTodo(c *gin.Context)
	CreateLabel(c *gin.Context)
	CreateComment(c *gin.Context)
}

type todosController struct {
	useCase Usecase
}

func NewTodosController(todoRouter *gin.RouterGroup, useCase Usecase) TodosController {
	controller := &todosController{
		useCase: useCase,
	}
	todoRouter.POST("", controller.CreateTodo)
	todoRouter.POST("/label", controller.CreateLabel)
	todoRouter.POST("/comment/:todo_id", controller.CreateComment)
	return controller
}

func (t *todosController) CreateComment(c *gin.Context) {
	todoId := c.Param("todo_id")
	var payload CreateCommentRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.Error(c, http.StatusInternalServerError, err)
		return
	}

	validationErr := validate.Struct(payload)
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

	err := t.useCase.CreateComment(payload, todoId)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithoutData(c, http.StatusCreated, "success create comment")
}

func (t *todosController) CreateLabel(c *gin.Context) {
	var payload CreateLabelRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.Error(c, http.StatusInternalServerError, err)
		return
	}

	validationErr := validate.Struct(payload)
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

	err := t.useCase.CreateLabel(payload)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithoutData(c, http.StatusCreated, "success create label")
}

func (t *todosController) CreateTodo(c *gin.Context) {
	var payload CreateTodoRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.Error(c, http.StatusInternalServerError, err)
		return
	}

	validationErr := validate.Struct(payload)
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

	err := t.useCase.CreateTodo(payload)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithoutData(c, http.StatusCreated, "success create todo")
}
