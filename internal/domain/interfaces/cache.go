package interfaces

import (
	"context"

	"github.com/unkabogaton/github-users/internal/domain/entities"
)

type Cache interface {
	GetUser(ctx context.Context, login string) (*entities.User, bool, error)
	SetUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, login string) error
}
