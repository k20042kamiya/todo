package repository

import (
	"context"

	"lambda/domain/entity"
)

type NotificationRepository interface {
	FindByTodoIDAndType(ctx context.Context, todoID int, notifType string) (*entity.Notification, error)
	Create(ctx context.Context, notification *entity.Notification) error
	FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error)
}
