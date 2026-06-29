package repository

import (
	"context"
	"errors"

	"notification/domain/entity"
	"notification/domain/repository"
	"notification/infrastructure/database"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetTx(ctx, r.db)
}

func (r *notificationRepository) FindByTodoIDAndType(ctx context.Context, todoID int, notifType string) (*entity.Notification, error) {
	var notification entity.Notification
	err := r.getDB(ctx).Where("todo_id = ? AND type = ?", todoID, notifType).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	return r.getDB(ctx).Create(notification).Error
}

func (r *notificationRepository) FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	err := r.getDB(ctx).
		Where("is_completed = ? AND due_date IS NOT NULL", false).
		Find(&todos).Error
	if err != nil {
		return nil, err
	}
	return todos, nil
}
