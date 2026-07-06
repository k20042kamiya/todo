package repository

import (
	"context"

	"notification/domain/entity"
	"notification/domain/repository"

	"gorm.io/gorm"
)

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) repository.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *todoRepository) FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	err := r.getDB(ctx).
		Where("is_completed = ? AND due_date IS NOT NULL", false).
		Find(&todos).Error
	if err != nil {
		return nil, err
	}
	return todos, nil
}
