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

func NewUserRepository(database *sqlx.DB) interfaces.UserRepository {
	genericRepository := NewGenericRepository[entities.User](database, "github_users", "id")
	return &UserRepository{GenericRepository: genericRepository}
}

func (userRepository *UserRepository) Upsert(
	ctx context.Context,
	userEntity *entities.User,
) error {
	return userRepository.GenericRepository.Upsert(ctx, *userEntity)
}

func (userRepository *UserRepository) BatchUpsert(
	ctx context.Context,
	userEntities *[]entities.User,
) error {
	return userRepository.GenericRepository.BatchUpsert(ctx, *userEntities)
}

func (userRepository *UserRepository) GetByLogin(
	ctx context.Context,
	login string,
) (*entities.User, error) {
	return userRepository.GetByField(ctx, "login", login)
}

func (userRepository *UserRepository) List(
	ctx context.Context,
	listOptions interfaces.ListOptions,
) ([]entities.User, error) {
	return userRepository.GenericRepository.List(
		ctx,
		listOptions.Limit,
		listOptions.Page,
		listOptions.OrderBy,
		listOptions.OrderDirection,
	)
}

func (userRepository *UserRepository) DeleteByLogin(
	ctx context.Context,
	login string,
) error {
	return userRepository.DeleteByField(ctx, "login", login)
}
