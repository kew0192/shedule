package auth

import (
	"context"
)

type UserLoginer interface {
	LoginUser(ctx context.Context, Code string) error
}
type AccessTokenValidator interface {
	AccessTokenValidate(ctx context.Context, refreshtoken string, id int) (string, error)
}
