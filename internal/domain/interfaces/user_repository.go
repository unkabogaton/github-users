package interfaces

import (
	"context"

	"github.com/unkabogaton/github-users/internal/domain/entities"
)

type UserRepository interface {
	Upsert(ctx context.Context, user *entities.User) error
	BatchUpsert(ctx context.Context, users *[]entities.User) error
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	List(ctx context.Context, options ListOptions) ([]entities.User, error)
	DeleteByLogin(ctx context.Context, login string) error
}
