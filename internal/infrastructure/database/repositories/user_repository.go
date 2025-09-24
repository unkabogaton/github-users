package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

type UserRepository struct {
	*GenericRepository[entities.User]
}

func NewUserRepository(db *sqlx.DB) interfaces.UserRepository {
	baseRepo := NewGenericRepository[entities.User](db, "github_users", "id")
	return &UserRepository{GenericRepository: baseRepo}
}

func (r *UserRepository) Upsert(ctx context.Context, u *entities.User) error {
	return r.GenericRepository.Upsert(ctx, *u)
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	return r.GetByField(ctx, "login", login)
}

func (r *UserRepository) List(ctx context.Context, options interfaces.ListOptions) ([]entities.User, error) {
    return r.GenericRepository.List(ctx, options.Limit, options.Page, options.OrderBy, options.OrderDirection)
}

func (r *UserRepository) DeleteByLogin(ctx context.Context, login string) error {
	return r.DeleteByField(ctx, "login", login)
}
