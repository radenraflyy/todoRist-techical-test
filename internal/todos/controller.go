package todos

import (
	"fmt"
	"net/http"
	_ "todorist/docs"
	"todorist/pkg/customlog"
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
	UpdateTodo(c *gin.Context)
	DeleteTodo(c *gin.Context)
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
	todoRouter.PATCH("", controller.UpdateTodo)
	todoRouter.DELETE("/:todo_id", controller.DeleteTodo)
	return controller
}

// CreateComment godoc
// @Summary     Tambah komentar ke todo
// @Description Endpoint ini menambahkan komentar pada todo tertentu
// @Tags        todos
// @Accept      json
// @Produce     json
// @Param       todo_id  path    string                       true  "ID todo"
// @Param  payload  body  todos.CreateCommentRequest  true  "Payload komentar"
// @Router      /todos/comment/{todo_id} [post]
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

// CreateLabel godoc
// @Summary     Buat label baru
// @Description Endpoint ini membuat satu label baru untuk todo
// @Tags        todos
// @Accept      json
// @Produce     json
// @Param       payload  body    todos.CreateLabelRequest  true  "Payload untuk membuat label"
// @Success     201      {object} todos.CreateLabelResponse
// @Router      /todos/label [post]
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

// CreateTodo godoc
// @Summary     Buat todo baru
// @Description Endpoint ini membuat satu todo baru
// @Tags        todos
// @Accept      json
// @Produce     json
// @Param       payload  body    todos.CreateTodoRequest  true  "Payload untuk membuat todo"
// @Router      /todos [post]
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

// GetAllLabels godoc
// @Summary     Daftar semua label
// @Description Mengambil semua label untuk user yang sedang login
// @Tags        todos
// @Produce     json
// @Success     200  {object} todos.GetAllLabelsResponse
// @Router      /todos/list-label [get]
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

// GetAllTodos godoc
// @Summary     Daftar semua todo
// @Description Mengambil daftar todo — sudah support pagination, search, filter
// @Tags        todos
// @Produce     json
// @Param       limit     query    int     false  "Limit per halaman"
// @Param       offset    query    int     false  "Halaman (1-based)"
// @Param       search    query    string  false  "Keyword pencarian"`
// @Param      status     query    string  false  "Filter status (true/false)"
// @Param       priority  query    string  false  "Filter prioritas"
// @Param       due_date  query    string  false  "Filter tanggal (YYYY-MM-DD)"
// @Success     200       {object} todos.GetAllTodosResponse
// @Router      /todos/list-todo [get]
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

	customlog.PrintJSON(filter, "filter AING")

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

// UpdateTodo godoc
// @Summary     Update status todo
// @Description Menandai satu atau banyak todo sebagai done/undone
// @Tags        todos
// @Accept      json
// @Produce     json
// @Param       payload  body    todos.UpdateTodoRequest  true  "Payload update status todo"
// @Router      /todos [patch]
func (t *todosController) UpdateTodo(c *gin.Context) {
	var payload UpdateTodoRequest
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
	customlog.PrintJSON(payload, "payload AING")

	err := t.useCase.UpdateTodo(payload)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithoutData(c, http.StatusOK, "success update todo")
}

// DeleteTodo godoc
// @Summary     Hapus todo
// @Description Menghapus satu todo berdasarkan ID
// @Tags        todos
// @Produce     json
// @Param       todo_id  path     string  true  "ID todo yang akan dihapus"
// @Router      /todos/{todo_id} [delete]
func (t *todosController) DeleteTodo(c *gin.Context) {
	todoId := c.Param("todo_id")
	err := t.useCase.DeleteTodo(todoId)
	if err != nil {
		c.Error(&exception.CustomException{
			Message: fmt.Sprintf("%v", err.Error()),
			Code:    http.StatusUnprocessableEntity,
		})
		return
	}

	utils.SuccessWithoutData(c, http.StatusOK, "success delete todo")
}
