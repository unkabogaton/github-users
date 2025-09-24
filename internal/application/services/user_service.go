package services

import (
	"context"

	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

type UserService struct {
	repository interfaces.UserRepository
	cache      interfaces.Cache
	client     interfaces.GitHubClient
}

func NewUserService(
	repository interfaces.UserRepository,
	cache interfaces.Cache,
	client interfaces.GitHubClient,
) interfaces.UserService {
	return &UserService{
		repository: repository,
		cache:      cache,
		client:     client,
	}
}

func (s *UserService) List(ctx context.Context, options interfaces.ListOptions) ([]entities.User, error) {
	if options.Limit <= 0 {
		options.Limit = 10
	}
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.OrderBy == "" {
		options.OrderBy = "id"
	}
	if options.OrderDirection == "" {
		options.OrderDirection = "ASC"
	}
	return s.repository.List(ctx, options)
}

func (s *UserService) Get(ctx context.Context, username string) (*entities.User, error) {
	if s.cache != nil {
		cachedUser, cacheHit, cacheError := s.cache.GetUser(ctx, username)
		if cacheError == nil && cacheHit {
			return cachedUser, nil
		}
	}

	ghUser, err := s.client.FetchOne(ctx, username)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		ID:        ghUser.ID,
		Login:     ghUser.Login,
		NodeID:    ghUser.NodeID,
		AvatarURL: ghUser.AvatarURL,
		HTMLURL:   ghUser.HTMLURL,
		URL:       ghUser.URL,
		Type:      ghUser.Type,
		SiteAdmin: ghUser.SiteAdmin,
	}

	if s.cache != nil {
		_ = s.cache.SetUser(ctx, user)
	}
	return user, nil
}

func (s *UserService) Update(
	ctx context.Context,
	username string,
	update interfaces.UpdateUserRequest,
) (*entities.User, error) {
	existingUser, getUserError := s.repository.GetByLogin(ctx, username)
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

	if upsertError := s.repository.Upsert(ctx, existingUser); upsertError != nil {
		return nil, upsertError
	}
	if s.cache != nil {
		_ = s.cache.SetUser(ctx, existingUser)
	}
	return existingUser, nil
}

func (s *UserService) Delete(ctx context.Context, username string) error {
	if deleteError := s.repository.DeleteByLogin(ctx, username); deleteError != nil {
		return deleteError
	}

	if s.cache != nil {
		_ = s.cache.DeleteUser(ctx, username)
	}
	return nil
}
