package user

import (
	"context"

	"todo/infrastructure/database"

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

func (r *repository) FindByID(ctx context.Context, id int) (*User, error) {
	var user User
	if err := r.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error) {
	var user User
	if err := r.getDB(ctx).Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.getDB(ctx).Create(user).Error
}
