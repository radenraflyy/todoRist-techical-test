package todos

import (
	"log"
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
	return t.db.Tx(func(tx *config.DB) error {
		data.UserId = t.db.GetUserId()

		responseTodo := struct {
			Id string `db:"id"`
		}{}

		if err := t.db.InsertOne(data, "todos", &responseTodo); err != nil {
			log.Println("error inserting todo:", err)
			return err
		}

		dataTablePivot := struct {
			TodoId  string `db:"todo_id"`
			LabelId string `db:"label_id"`
		}{
			TodoId:  responseTodo.Id,
			LabelId: data.LabelId,
		}

		if err := t.db.InsertOne(dataTablePivot, "todo_label_pivot", nil); err != nil {
			log.Println("error inserting pivot:", err)
			return err
		}

		return nil
	})
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
