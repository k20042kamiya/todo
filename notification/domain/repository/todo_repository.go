package repository

import (
	"context"

	"notification/domain/entity"
)

type TodoRepository interface {
	FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error)
}
