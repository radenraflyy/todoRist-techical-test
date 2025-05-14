package todos

import (
	"todorist/config"
)

type TodosRepository interface {
	CreateTodo(data CreateTodoRequest) error
	CreateLabel(data CreateLabelRequest) error
	CreateComment(data CreateCommentRequest, todoId string) error
}

type todosRepository struct {
	db *config.DB
}

func NewTodosRepository(db *config.DB) TodosRepository {
	return &todosRepository{db}
}

func (t *todosRepository) CreateTodo(data CreateTodoRequest) error {
	data.UserId = t.db.GetUserId()
	if err := t.db.InsertOne(data, "todos", nil); err != nil {
		return err
	}
	return nil
}

func (t *todosRepository) CreateLabel(data CreateLabelRequest) error {
	data.UserId = t.db.GetUserId()
	if err := t.db.InsertOne(data, "label_todos", nil); err != nil {
		return err
	}
	return nil
}

func (t *todosRepository) CreateComment(data CreateCommentRequest, todoId string) error {
	data.TodoId = todoId
	if err := t.db.InsertOne(data, "comments", nil); err != nil {
		return err
	}
	return nil
}
