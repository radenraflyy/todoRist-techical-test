// request.dto.go
package todos

type (
	CreateTodoRequest struct {
		Title       string   `json:"title" db:"title" validate:"required"`
		Description string   `json:"description" db:"description" validate:"required"`
		DueDate     string   `json:"due_date" db:"due_date" validate:"required"`
		IsDone      bool     `json:"is_done" db:"is_done"`
		Priority    string   `json:"priority" db:"priority"`
		LabelIds    []string `json:"label"`
		UserId      string   `json:"user_id" db:"user_id"`
	}

	CreateLabelRequest struct {
		Name   string `json:"name" db:"name" validate:"required"`
		UserId string `json:"user_id" db:"user_id"`
	}

	CreateCommentRequest struct {
		TodoId  string `json:"todo_id" db:"todo_id"`
		Comment string `json:"comment" db:"comment" validate:"required"`
	}

	FilteringTodosRequest struct {
		Status   string `form:"status"`
		Priority string `form:"priority"`
		DueDate  string `form:"due_date"`
		OrderBy  string `form:"order_by"`
		Order    string `form:"order"`
		Search   string `form:"search"`
		Limit    int    `form:"limit"`
		Offset   int    `form:"offset"`
	}

	UpdateTodoRequest struct {
		TodoId []string `json:"id" db:"id"`
		IsDone bool     `json:"is_done" db:"is_done"`
	}
)
