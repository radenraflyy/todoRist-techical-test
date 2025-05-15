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
)
