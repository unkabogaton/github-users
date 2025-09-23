package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/unkabogaton/github-users/internal/models"
)

type UserRepo interface {
	Upsert(ctx context.Context, u *models.User) error
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	DeleteByLogin(ctx context.Context, login string) error
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
            type, user_view_type, site_admin
        ) VALUES (
            :id, :login, :node_id, :avatar_url, :url, :html_url,
            :type, :user_view_type, :site_admin
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

func (r *sqlxUserRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var u models.User
	query := `SELECT id, login, node_id, avatar_url, url, html_url,
                     type, user_view_type, site_admin, updated_at, created_at
              FROM github_users WHERE login = $1 LIMIT 1`
	if err := r.db.GetContext(ctx, &u, query, login); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *sqlxUserRepo) List(ctx context.Context) ([]models.User, error) {
	var list []models.User
	query := `SELECT id, login, node_id, avatar_url, url, html_url,
                     type, user_view_type, site_admin, updated_at, created_at
              FROM github_users`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *sqlxUserRepo) DeleteByLogin(ctx context.Context, login string) error {
	query := `DELETE FROM github_users WHERE login = $1`
	_, err := r.db.ExecContext(ctx, query, login)
	return err
}
