package user

import (
	"context"

	apperrors "todo/shared/errors"
)

type Usecase interface {
	FindOrCreateByFirebaseUID(ctx context.Context, firebaseUID, email, name string) (*User, error)
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) FindOrCreateByFirebaseUID(ctx context.Context, firebaseUID, email, name string) (*User, error) {
	// 既存ユーザーは最初のFindで高速に返す（大多数のリクエスト）
	user, err := u.repo.FindByFirebaseUID(ctx, firebaseUID)
	if err == nil {
		return user, nil
	}
	if apperrors.GetCode(err) != apperrors.ErrCodeNotFound {
		return nil, err
	}

	if name == "" {
		name = "Unknown"
	}

	// INSERT IGNORE + SELECT で競合状態を回避
	user = &User{
		FirebaseUID: firebaseUID,
		Email:       email,
		Name:        name,
	}
	if err := u.repo.FindOrCreate(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
