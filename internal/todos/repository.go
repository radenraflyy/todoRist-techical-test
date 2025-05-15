package todos

import (
	"log"
	"todorist/config"
)

type TodosRepository interface {
	CreateTodo(data CreateTodoRequest) error
	CreateLabel(data CreateLabelRequest) (CreateLabelResponse, error)
	CreateComment(data CreateCommentRequest, todoId string) error
	GetAllLabels(userId string) ([]GetAllLabelsResponse, error)
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

		var dataTablePivot []struct {
			TodoId  string `db:"todo_id"`
			LabelId string `db:"label_id"`
		}

		for _, labelId := range data.LabelIds {
			dataTablePivot = append(dataTablePivot, struct {
				TodoId  string `db:"todo_id"`
				LabelId string `db:"label_id"`
			}{
				TodoId:  responseTodo.Id,
				LabelId: labelId,
			})
		}

		if err := t.db.InsertMany(dataTablePivot, "todo_label_pivot", nil); err != nil {
			log.Println("error inserting pivot:", err)
			return err
		}

		return nil
	})
}

func (t *todosRepository) CreateLabel(data CreateLabelRequest) (CreateLabelResponse, error) {
	data.UserId = t.db.GetUserId()
	var CreateLabelResponse CreateLabelResponse
	if err := t.db.InsertOne(data, "label_todos", &CreateLabelResponse); err != nil {
		return CreateLabelResponse, err
	}
	return CreateLabelResponse, nil
}

func (t *todosRepository) CreateComment(data CreateCommentRequest, todoId string) error {
	data.TodoId = todoId
	if err := t.db.InsertOne(data, "comments", nil); err != nil {
		return err
	}
	return nil
}

func (t *todosRepository) GetAllLabels(userId string) ([]GetAllLabelsResponse, error) {
	data := make([]GetAllLabelsResponse, 0)
	query := `SELECT id, name FROM label_todos WHERE user_id = $<user_id>`
	if err := t.db.SelectMany(query, &data, map[string]any{"user_id": userId}); err != nil {
		log.Println("error getting all labels:", err)
	}

	return data, nil
}
