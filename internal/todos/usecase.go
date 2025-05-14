package todos

import "todorist/config"

type Usecase interface {
	CreateTodo(data CreateTodoRequest) error
	CreateLabel(data CreateLabelRequest) error
	CreateComment(data CreateCommentRequest, todoId string) error
}

type useCase struct {
	repo TodosRepository
	db   *config.DB
}

func NewUseCase(repo TodosRepository, db *config.DB) Usecase {
	return &useCase{
		repo: repo,
		db:   db,
	}
}

func (u *useCase) CreateComment(data CreateCommentRequest, todoId string) error {
	if err := u.repo.CreateComment(data, todoId); err != nil {
		return err
	}
	return nil
}

func (u *useCase) CreateLabel(data CreateLabelRequest) error {
	if err := u.repo.CreateLabel(data); err != nil {
		return err
	}
	return nil
}

func (u *useCase) CreateTodo(data CreateTodoRequest) error {
	if err := u.repo.CreateTodo(data); err != nil {
		return err
	}
	return nil
}
