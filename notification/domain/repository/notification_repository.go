package repository

import (
	"context"

	"notification/domain/entity"
)

type NotificationRepository interface {
	FindTodayByTodoID(ctx context.Context, todoID int) (*entity.Notification, error)
	Create(ctx context.Context, notification *entity.Notification) error
}
