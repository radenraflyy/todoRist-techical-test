package todos

import (
	"fmt"
	"log"
	"strings"
	"todorist/config"
)

type TodosRepository interface {
	CreateTodo(data CreateTodoRequest) error
	CreateLabel(data CreateLabelRequest) (CreateLabelResponse, error)
	CreateComment(data CreateCommentRequest, todoId string) error
	GetAllLabels(userId string) ([]GetAllLabelsResponse, error)
	GetAllTodos(userId string, filter FilteringTodosRequest) ([]GetAllTodosResponse, error)
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

func (t *todosRepository) GetAllTodos(userId string, q FilteringTodosRequest) ([]GetAllTodosResponse, error) {
	data := make([]GetAllTodosResponse, 0)
	limit := q.Limit
	if limit <= 0 {
		limit = 5
	}
	offset := (q.Offset - 1) * limit
	if offset < 0 {
		offset = 0
	}

	allowedSortBy := map[string]bool{"title": true, "due_date": true, "created_at": true, "priority": true}
	allowedOrder := map[string]bool{"asc": true, "desc": true}

	orderBy := q.OrderBy
	if !allowedSortBy[orderBy] {
		orderBy = "created_at"
	}

	order := q.Order
	if !allowedOrder[order] {
		order = "desc"
	}

	search := q.Search

	wherearr := make([]string, 0)
	wherearr = append(wherearr, "todos.user_id = $<user_id>")
	wherearr = append(wherearr, "todos.is_done = false")
	wherearr = append(wherearr, "todos.deleted_at IS NULL")

	if search != "" {
		wherearr = append(wherearr, `(LOWER(todos.title) LIKE LOWER($<search>) OR LOWER(todos.description) LIKE LOWER($<search>))`)
	}

	wherestr := strings.Join(wherearr, " AND ")

	query := fmt.Sprintf(`
	SELECT
		COUNT(*) OVER () AS count,
		todos.id,
		todos.title,
		todos.description,
		todos.created_at,
		todos.due_date,
		todos.priority,
		todos.is_done
	FROM todos
	WHERE %s
	ORDER BY $<orderBy:raw> $<order:raw> NULLS LAST, todos.created_at DESC
	LIMIT $<limit>
	OFFSET $<offset>
	`, wherestr)

	params := map[string]interface{}{
		"user_id": userId,
		"search":  "%" + search + "%",
		"orderBy": orderBy,
		"order":   order,
		"limit":   limit,
		"offset":  offset,
	}

	err := t.db.SelectMany(query, &data, params)
	if err != nil {
		log.Println("error getting all todos:", err)
		return nil, err
	}

	return data, nil
}
