package todos

import (
	"todorist/config"
)

type Usecase interface {
	CreateTodo(data CreateTodoRequest) error
	CreateLabel(data CreateLabelRequest) (CreateLabelResponse, error)
	CreateComment(data CreateCommentRequest, todoId string) error
	GetAllLabels(userId string) ([]GetAllLabelsResponse, error)
	GetAllTodos(userId string, filter FilteringTodosRequest) ([]GetAllTodosResponse, error)
	UpdateTodo(data UpdateTodoRequest) error
	DeleteTodo(todoId string) error
	GetDetailTodo(todoId string) (GetDetailTodosResponse, error)
	UpdateTaskTodo(todoId string, data UpdateDetailTodo) error
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

func (u *useCase) CreateLabel(data CreateLabelRequest) (CreateLabelResponse, error) {
	resp, err := u.repo.CreateLabel(data)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (u *useCase) CreateTodo(data CreateTodoRequest) error {
	if err := u.repo.CreateTodo(data); err != nil {
		return err
	}
	return nil
}

func (u *useCase) GetAllLabels(userId string) ([]GetAllLabelsResponse, error) {
	resp, err := u.repo.GetAllLabels(userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (u *useCase) GetAllTodos(userId string, filter FilteringTodosRequest) ([]GetAllTodosResponse, error) {
	resp, err := u.repo.GetAllTodos(userId, filter)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (t *useCase) UpdateTodo(data UpdateTodoRequest) error {
	if err := t.repo.UpdateTodoMany(data); err != nil {
		return err
	}
	return nil
}

func (u *useCase) DeleteTodo(todoId string) error {
	if err := u.repo.DeleteTodo(todoId); err != nil {
		return err
	}
	return nil
}

func (u *useCase) GetDetailTodo(todoId string) (GetDetailTodosResponse, error) {
	resp, err := u.repo.GetDetailTodo(todoId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (u *useCase) UpdateTaskTodo(todoId string, data UpdateDetailTodo) error {
	if err := u.repo.UpdateTaskTodo(todoId, data); err != nil {
		return err
	}
	return nil
}
