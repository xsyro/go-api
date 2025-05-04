package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/xsyro/goapi/internal/app/api/requests"
	"github.com/xsyro/goapi/internal/app/repo"
	model "github.com/xsyro/goapi/internal/app/repo/sqlc"
	"github.com/xsyro/goapi/internal/utils"
)

type todoService struct {
	repo repo.Repository
}

func NewTodoService(repo repo.Repository) *todoService {
	return &todoService{repo: repo}
}

func (s *todoService) GetTodos(ctx context.Context) ([]model.Todo, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	todos, err := s.repo.GetTodos(ctx)
	if err != nil {
		return []model.Todo{}, repo.WrapDbError(err)
	}
	return todos, nil
}

func (s *todoService) GetTodo(ctx context.Context, id uuid.UUID) (model.Todo, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	todo, err := s.repo.GetTodo(ctx, id)
	if err != nil {
		return model.Todo{}, repo.WrapDbError(err)
	}
	return todo, nil

}

func (s *todoService) CreateTodo(ctx context.Context, tr requests.PostTodo) (model.Todo, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	todo, err := s.repo.CreateTodo(ctx, model.CreateTodoParams{
		UserID: uuid.MustParse(tr.UserID),
		Task:   tr.Task,
	})
	if err != nil {
		return model.Todo{}, repo.WrapDbError(err)
	}
	return todo, nil
}

func (s *todoService) UpdateTodo(ctx context.Context, tr requests.PatchTodo, id uuid.UUID) error {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	err := s.repo.UpdateTodo(ctx, model.UpdateTodoParams{
		ID:   id,
		Task: tr.Task,
		Done: tr.Done,
	})
	if err != nil {
		return repo.WrapDbError(err)
	}
	return nil
}

func (s *todoService) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	err := s.repo.DeleteTodo(ctx, id)
	if err != nil {
		return repo.WrapDbError(err)
	}
	return nil

}
