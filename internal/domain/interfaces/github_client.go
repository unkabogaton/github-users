package interfaces

import (
	"context"

	"github.com/unkabogaton/github-users/internal/domain/entities"
)

type GitHubClient interface {
	FetchUsersSince(ctx context.Context, lastUserID, resultsPerPage int) ([]entities.GitHubUser, error)
	FetchOne(ctx context.Context, username string) (*entities.GitHubUser, error)
}
