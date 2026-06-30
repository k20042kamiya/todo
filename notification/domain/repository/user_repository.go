package repository

import (
	"context"

	"notification/domain/entity"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*entity.User, error)
}
