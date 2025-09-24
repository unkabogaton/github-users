package interfaces

import (
	"context"

	"github.com/unkabogaton/github-users/internal/domain/entities"
)

type UserService interface {
	Get(ctx context.Context, username string) (*entities.User, error)
	List(ctx context.Context) ([]entities.User, error)
	Update(ctx context.Context, username string, update UpdateUserRequest) (*entities.User, error)
	Delete(ctx context.Context, username string) error
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
