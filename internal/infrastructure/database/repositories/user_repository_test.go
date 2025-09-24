package repositories

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

var sampleUser = &entities.User{
	ID:           1,
	Login:        "sample_username",
	NodeID:       "N1",
	AvatarURL:    "http//",
	URL:          "http//",
	HTMLURL:      "http//",
	Type:         "User",
	UserViewType: "public",
	SiteAdmin:    false,
	UpdatedAt:    time.Now(),
	CreatedAt:    time.Now(),
}

func TestUserRepository_Upsert(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repository := NewUserRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO github_users")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repository.Upsert(context.Background(), sampleUser)
	require.NoError(t, err)
}

func TestUserRepository_List_WithPaginationAndOrdering(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repository := NewUserRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"id", "login", "node_id", "avatar_url", "url", "html_url", "type", "user_view_type", "site_admin", "updated_at", "created_at",
	}).AddRow(
		sampleUser.ID, sampleUser.Login, sampleUser.NodeID, sampleUser.AvatarURL, sampleUser.URL, sampleUser.HTMLURL,
		sampleUser.Type, sampleUser.UserViewType, sampleUser.SiteAdmin, sampleUser.UpdatedAt, sampleUser.CreatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, login, node_id, avatar_url, url, html_url, type, user_view_type, site_admin, updated_at, created_at FROM github_users ORDER BY login ASC LIMIT ? OFFSET ?",
	)).
		WithArgs(5, 5).
		WillReturnRows(rows)

	options := interfaces.ListOptions{Limit: 5, Page: 2, OrderBy: "login", OrderDirection: "ASC"}
	users, err := repository.List(context.Background(), options)
	require.NoError(t, err)
	require.Len(t, users, 1)
}

func TestUserRepository_GetByLogin(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repository := NewUserRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"id", "login", "node_id", "avatar_url", "url", "html_url",
		"type", "user_view_type", "site_admin", "updated_at", "created_at",
	}).AddRow(
		sampleUser.ID, sampleUser.Login, sampleUser.NodeID, sampleUser.AvatarURL, sampleUser.URL, sampleUser.HTMLURL,
		sampleUser.Type, sampleUser.UserViewType, sampleUser.SiteAdmin, sampleUser.UpdatedAt, sampleUser.CreatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, login, node_id, avatar_url, url, html_url, type, user_view_type, site_admin, updated_at, created_at FROM github_users WHERE login = ? LIMIT 1",
	)).
		WithArgs(sampleUser.Login).
		WillReturnRows(rows)

	got, err := repository.GetByLogin(context.Background(), "sample_username")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, sampleUser.ID, got.ID)
	require.Equal(t, sampleUser.Login, got.Login)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_DeleteByLogin(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repository := NewUserRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM github_users WHERE login = ?",
	)).
		WithArgs("sample_username").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repository.DeleteByLogin(context.Background(), "sample_username")
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByLogin_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repository := NewUserRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, login, node_id, avatar_url, url, html_url, type, user_view_type, site_admin, updated_at, created_at FROM github_users WHERE login = ? LIMIT 1",
	)).
		WithArgs("nonexistent_user").
		WillReturnError(fmt.Errorf("no row found"))

	user, err := repository.GetByLogin(context.Background(), "nonexistent_user")
	require.Error(t, err)
	require.Nil(t, user)
}
