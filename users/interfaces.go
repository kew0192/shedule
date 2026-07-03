package users

import "context"

type UserCreator interface {
	CreateUser(ctx context.Context, user *User) error
}

type UserDeleter interface {
	DeleteUser(ctx context.Context, id int) error
}
type UserFinder interface {
	FindUser(ctx context.Context, id int) (*User, error)
}
