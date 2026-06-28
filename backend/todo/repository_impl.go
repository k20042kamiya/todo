package todo

import (
	"context"
	"errors"
	"log"

	"todo/infrastructure/database"
	apperrors "todo/shared/errors"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB(ctx context.Context) *gorm.DB {
	return database.GetTx(ctx, r.db)
}

func (r *repository) FindByUserID(ctx context.Context, userID int) ([]*Todo, error) {
	var todos []*Todo
	if err := r.getDB(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error; err != nil {
		log.Printf("[WARN] FindByUserID failed: userID=%d, error=%v", userID, err)
		return nil, err
	}
	return todos, nil
}

func (r *repository) FindByID(ctx context.Context, id int) (*Todo, error) {
	var todo Todo
	if err := r.getDB(ctx).Where("id = ?", id).First(&todo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "todo not found")
		}
		return nil, err
	}
	return &todo, nil
}

func (r *repository) Create(ctx context.Context, todo *Todo) error {
	return r.getDB(ctx).Create(todo).Error
}

func (r *repository) Update(ctx context.Context, todo *Todo) error {
	return r.getDB(ctx).Save(todo).Error
}

func (r *repository) Delete(ctx context.Context, id int, userID int) error {
	return r.getDB(ctx).Where("user_id = ?", userID).Delete(&Todo{ID: id}).Error
}
