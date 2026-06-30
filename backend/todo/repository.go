package todo

import "context"

type Repository interface {
	FindByUserID(ctx context.Context, userID int) ([]*Todo, error)
	FindByID(ctx context.Context, id int) (*Todo, error)
	Create(ctx context.Context, todo *Todo) error
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, id int, userID int) error
}
