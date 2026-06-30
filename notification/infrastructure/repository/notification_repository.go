package repository

import (
	"context"
	"errors"
	"time"

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

func (r *notificationRepository) FindTodayByTodoID(ctx context.Context, todoID int) (*entity.Notification, error) {
	var notification entity.Notification
	today := time.Now().UTC().Truncate(24 * time.Hour)
	err := r.getDB(ctx).
		Where("todo_id = ? AND sent_at >= ?", todoID, today).
		First(&notification).Error
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

