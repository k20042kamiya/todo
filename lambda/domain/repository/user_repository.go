package repository

import (
	"context"

	"lambda/domain/entity"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*entity.User, error)
}
