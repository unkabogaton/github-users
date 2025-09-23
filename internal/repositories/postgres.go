package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/unkabogaton/github-users/internal/models"
)

type UserRepo interface {
	Upsert(ctx context.Context, u *models.User) error
}

type sqlxUserRepo struct {
	db *sqlx.DB
}

func NewSQLXUserRepo(db *sqlx.DB) UserRepo {
	return &sqlxUserRepo{db: db}
}

func (r *sqlxUserRepo) Upsert(ctx context.Context, u *models.User) error {
	query := `
        INSERT INTO github_users (
            id, login, node_id, avatar_url, url, html_url,
            type, user_view_type, site_admin, created_at, updated_at
        ) VALUES (
            :id, :login, :node_id, :avatar_url, :url, :html_url,
            :type, :user_view_type, :site_admin, COALESCE(:created_at, NOW()), NOW()
        )
        ON CONFLICT (id) DO UPDATE SET
            login          = EXCLUDED.login,
            node_id        = EXCLUDED.node_id,
            avatar_url     = EXCLUDED.avatar_url,
            url            = EXCLUDED.url,
            html_url       = EXCLUDED.html_url,
            type           = EXCLUDED.type,
            user_view_type = EXCLUDED.user_view_type,
            site_admin     = EXCLUDED.site_admin,
            updated_at     = NOW()
    `
	_, err := r.db.NamedExecContext(ctx, query, u)
	return err
}
