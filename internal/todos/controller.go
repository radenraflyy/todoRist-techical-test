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
	GetAllLabels(c *gin.Context)
	GetAllTodos(c *gin.Context)
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
	todoRouter.GET("/list-label", controller.GetAllLabels)
	todoRouter.GET("/list-todo", controller.GetAllTodos)
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

	resp, err := t.useCase.CreateLabel(payload)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithData(c, http.StatusCreated, resp, "success create label")
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

func (t *todosController) GetAllLabels(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.Error(&exception.CustomException{
			Message: "user id not found",
			Code:    http.StatusNotFound,
		})
		return
	}
	res, err := t.useCase.GetAllLabels(userId.(string))
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithData(c, http.StatusOK, res, "success get all labels")
}

func (t *todosController) GetAllTodos(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.Error(&exception.CustomException{
			Message: "user id not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	var filter FilteringTodosRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 5
	}
	if filter.Offset == 0 {
		filter.Offset = 0
	}
	if filter.OrderBy == "" {
		filter.OrderBy = "created_at"
	}
	if filter.Order == "" {
		filter.Order = "desc"
	}

	res, err := t.useCase.GetAllTodos(userId.(string), filter)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	totalItems := 0
	if len(res) > 0 {
		totalItems = res[0].Count
	}

	data := struct {
		Items      []GetAllTodosResponse `json:"items"`
		TotalItems int                   `json:"totalItems"`
		Page       int                   `json:"page"`
		PerPage    int                   `json:"perPage"`
	}{
		Items:      res,
		TotalItems: totalItems,
		Page:       filter.Offset,
		PerPage:    filter.Limit,
	}

	utils.SuccessWithData(c, http.StatusOK, data, "success get all todos")
}
