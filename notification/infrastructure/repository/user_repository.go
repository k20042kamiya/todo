package repository

import (
	"context"
	"errors"
	"fmt"

	"notification/domain/entity"
	"notification/domain/repository"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	if err := r.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: id=%d: %w", id, entity.ErrNotFound)
		}
		return nil, fmt.Errorf("FindByID failed: id=%d: %w", id, err)
	}
	return &user, nil
}
