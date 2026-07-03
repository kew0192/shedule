package users

import (
	"context"
)

type Repository interface {
	SaveUserDB(ctx context.Context, user *User) error
	DeleteUserDB(ctx context.Context, id int) error
	FindUserDB(ctx context.Context, id int) (*User, error)
	FindUserCodeDB(ctx context.Context, code string) (*User, error)
	UpdateUserDB(ctx context.Context, user *User) error
	GetAllUsersDB(ctx context.Context) ([]*User, error)
}
