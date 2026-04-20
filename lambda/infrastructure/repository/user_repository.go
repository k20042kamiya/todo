package repository

import (
	"context"

	"lambda/domain/entity"
	"lambda/domain/repository"
	"lambda/infrastructure/database"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetTx(ctx, r.db)
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	if err := r.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
