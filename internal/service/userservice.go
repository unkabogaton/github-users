package service

import (
	"context"

	"github.com/unkabogaton/github-users/internal/cache"
	"github.com/unkabogaton/github-users/internal/models"
	"github.com/unkabogaton/github-users/internal/repositories"
)

type UserService struct {
	repository repositories.UserRepo
	cache      *cache.RedisCache
}

func NewUserService(repository repositories.UserRepo, cache *cache.RedisCache) *UserService {
	return &UserService{repository, cache}
}

func (service *UserService) List(ctx context.Context) ([]models.User, error) {
	return service.repository.List(ctx)
}
