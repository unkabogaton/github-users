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

type UpdateUserRequest struct {
	Login        string `json:"Login"`
	NodeID       string `json:"NodeID"`
	AvatarURL    string `json:"AvatarURL"`
	URL          string `json:"URL"`
	HTMLURL      string `json:"HTMLURL"`
	Type         string `json:"Type"`
	UserViewType string `json:"UserViewType"`
	SiteAdmin    bool   `json:"SiteAdmin"`
}

func (service *UserService) Update(
	ctx context.Context,
	username string,
	update UpdateUserRequest,
) (*models.User, error) {
	existingUser, getUserError := service.repository.GetByLogin(ctx, username)
	if getUserError != nil {
		return nil, getUserError
	}
	if update.Login != "" {
		existingUser.Login = update.Login
	}
	if update.NodeID != "" {
		existingUser.NodeID = update.NodeID
	}
	if update.AvatarURL != "" {
		existingUser.AvatarURL = update.AvatarURL
	}
	if update.URL != "" {
		existingUser.URL = update.URL
	}
	if update.HTMLURL != "" {
		existingUser.HTMLURL = update.HTMLURL
	}
	if update.Type != "" {
		existingUser.Type = update.Type
	}
	if update.UserViewType != "" {
		existingUser.UserViewType = update.UserViewType
	}
	existingUser.SiteAdmin = update.SiteAdmin

	if upsertError := service.repository.Upsert(ctx, existingUser); upsertError != nil {
		return nil, upsertError
	}
	_ = service.cache.SetUser(ctx, existingUser)
	return existingUser, nil
}

func (service *UserService) Delete(ctx context.Context, username string) error {
	if deleteError := service.repository.DeleteByLogin(ctx, username); deleteError != nil {
		return deleteError
	}

	if service.cache != nil {
		_ = service.cache.DeleteUser(ctx, username)
	}
	return nil
}

func (service *UserService) Get(
	requestContext context.Context,
	username string,
) (*models.User, error) {
	if service.cache != nil {
		cachedUser, cacheHit, cacheError := service.cache.GetUser(requestContext, username)
		if cacheError == nil && cacheHit {
			return cachedUser, nil
		}
	}

	databaseUser, databaseError := service.repository.GetByLogin(requestContext, username)
	if databaseError != nil {
		return nil, databaseError
	}

	_ = service.cache.SetUser(requestContext, databaseUser)
	return databaseUser, nil
}
