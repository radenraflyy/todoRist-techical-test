// response.dto.go
package todos

type (
	GetAllLabelsResponse struct {
		Id   string `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
	}

	CreateLabelResponse struct {
		Id   string `db:"id"`
		Name string `db:"name"`
	}

	GetAllTodosResponse struct {
		Count       int    `json:"-" db:"count"`
		Id          string `json:"id" db:"id"`
		Title       string `json:"title" db:"title"`
		Description string `json:"description" db:"description"`
		DueDate     string `json:"due_date" db:"due_date"`
		IsDone      bool   `json:"is_done" db:"is_done"`
		Priority    string `json:"priority" db:"priority"`
		CreatedAt   string `json:"created_at" db:"created_at"`
	}

	CommentResponse struct {
		Comment   string `json:"comment" db:"comment"`
		CreatedAt string `json:"created_at" db:"created_at"`
	}

	ResponseLable struct {
		Id   string `db:"id"`
		Name string `db:"name"`
	}

	GetDetailTodosResponse struct {
		Name            string            `json:"name" db:"name"`
		Title           string            `json:"title" db:"title"`
		Description     string            `json:"description" db:"description"`
		DueDate         string            `json:"due_date" db:"due_date"`
		IsDone          bool              `json:"is_done" db:"is_done"`
		Priority        string            `json:"priority" db:"priority"`
		ResponseLable   []ResponseLable   `json:"label"`
		CommentResponse []CommentResponse `json:"comment"`
	}
)
