package user

import (
	"context"
	"errors"

	"todo/infrastructure/database"
	apperrors "todo/shared/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *repository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error) {
	var user User
	if err := r.getDB(ctx).Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "user not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrCodeDatabase, "FindByFirebaseUID", err)
	}
	return &user, nil
}

func (r *repository) FindOrCreate(ctx context.Context, user *User) error {
	result := r.getDB(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "firebase_uid"}},
			DoNothing: true,
		}).
		Create(user)
	if result.Error != nil {
		return apperrors.Wrap(apperrors.ErrCodeDatabase, "FindOrCreate user", result.Error)
	}
	if result.RowsAffected == 0 {
		// 競合して INSERT がスキップされた場合、既存レコードを取得
		if err := r.getDB(ctx).Where("firebase_uid = ?", user.FirebaseUID).First(user).Error; err != nil {
			return apperrors.Wrap(apperrors.ErrCodeDatabase, "FindOrCreate fetch existing", err)
		}
	}
	return nil
}
