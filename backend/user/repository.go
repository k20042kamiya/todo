package user

import "context"

type Repository interface {
	FindByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error)
	FindOrCreate(ctx context.Context, user *User) error
}
