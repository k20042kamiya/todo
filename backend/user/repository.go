package user

import "context"

type Repository interface {
	FindByID(ctx context.Context, id int) (*User, error)
	FindByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error)
	Create(ctx context.Context, user *User) error
}
