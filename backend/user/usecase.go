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

	user = &User{
		FirebaseUID: firebaseUID,
		Email:       email,
		Name:        name,
	}
	if err := u.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
