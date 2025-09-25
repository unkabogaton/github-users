package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

type fakeRepository struct {
	stored map[string]*entities.User
}

func (f *fakeRepository) Upsert(ctx context.Context, user *entities.User) error {
	f.stored[user.Login] = user
	return nil
}

func (f *fakeRepository) BatchUpsert(ctx context.Context, users *[]entities.User) error {
	for _, user := range *users {
		f.stored[user.Login] = &user
	}
	return nil
}

func (f *fakeRepository) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	if u, ok := f.stored[login]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeRepository) List(ctx context.Context, options interfaces.ListOptions) ([]entities.User, error) {
	var out []entities.User
	for _, u := range f.stored {
		out = append(out, *u)
	}
	return out, nil
}
func (f *fakeRepository) DeleteByLogin(ctx context.Context, login string) error {
	delete(f.stored, login)
	return nil
}

type fakeCache struct{ items map[string]*entities.User }

func (f *fakeCache) GetUser(ctx context.Context, login string) (*entities.User, bool, error) {
	u, ok := f.items[login]
	return u, ok, nil
}
func (f *fakeCache) SetUser(ctx context.Context, user *entities.User) error {
	f.items[user.Login] = user
	return nil
}
func (f *fakeCache) DeleteUser(ctx context.Context, login string) error {
	delete(f.items, login)
	return nil
}

type fakeGitHubClient struct{}

func (f *fakeGitHubClient) FetchUsersSince(ctx context.Context, lastUserID int, resultsPerPage int) ([]entities.GitHubUser, error) {
	return nil, nil
}
func (f *fakeGitHubClient) FetchOne(ctx context.Context, username string) (*entities.GitHubUser, error) {
	return &entities.GitHubUser{ID: 1, Login: username}, nil
}

func TestUserService_Get_UsesCacheThenClient(t *testing.T) {
	t.Parallel()
	repo := &fakeRepository{stored: map[string]*entities.User{}}
	cache := &fakeCache{items: map[string]*entities.User{"octo": {ID: 1, Login: "octo"}}}
	client := &fakeGitHubClient{}
	svc := NewUserService(repo, cache, client)

	u, err := svc.Get(context.Background(), "octo")
	require.NoError(t, err)
	require.Equal(t, "octo", u.Login)
}

func TestUserService_Update_PersistsAndCaches(t *testing.T) {
	t.Parallel()
	repo := &fakeRepository{stored: map[string]*entities.User{"octo": {ID: 1, Login: "octo"}}}
	cache := &fakeCache{items: map[string]*entities.User{}}
	client := &fakeGitHubClient{}
	svc := NewUserService(repo, cache, client)

	updated, err := svc.Update(context.Background(), "octo", interfaces.UpdateUserRequest{Login: "octo"})
	require.NoError(t, err)
	require.Equal(t, "octo", updated.Login)
}

func TestUserService_List_Defaults(t *testing.T) {
	t.Parallel()
	repo := &fakeRepository{stored: map[string]*entities.User{"a": {ID: 1, Login: "a"}}}
	cache := &fakeCache{items: map[string]*entities.User{}}
	client := &fakeGitHubClient{}
	svc := NewUserService(repo, cache, client)

	users, err := svc.List(context.Background(), interfaces.ListOptions{})
	require.NoError(t, err)
	require.Len(t, users, 1)
}
