package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

type fakeUserService struct{}

func (f *fakeUserService) Get(ctx context.Context, username string) (*entities.User, error) {
	return &entities.User{ID: 1, Login: username}, nil
}

func (f *fakeUserService) List(ctx context.Context, _ interfaces.ListOptions) ([]entities.User, error) {
	return []entities.User{{ID: 1, Login: "sample_username"}}, nil
}

func (f *fakeUserService) Update(ctx context.Context, username string, update interfaces.UpdateUserRequest) (*entities.User, error) {
	return &entities.User{ID: 1, Login: update.Login}, nil
}

func (f *fakeUserService) Delete(ctx context.Context, username string) error { return nil }

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	controller := NewUserController(&fakeUserService{})
	router.GET("/users", controller.ListUsers)
	router.GET("/users/:username", controller.GetUser)
	router.PUT("/users/:username", controller.UpdateUser)
	router.DELETE("/users/:username", controller.DeleteUser)
	return router
}

func TestListUsers_OK(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestListUsers_InvalidLimit(t *testing.T) {
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodGet,
		"/users?limit=notANumber&page=2&orderby=login&order=desc", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestListUsers_InvalidPage(t *testing.T) {
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodGet,
		"/users?limit=10&page=notANumber&orderby=login&order=desc", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetUser_OK(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodGet, "/users/sample_username", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestUpdateUser_OK(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	updateBody := `{"Login":"sample_username"}`
	request := httptest.NewRequest(http.MethodPut,
		"/users/sample_username", strings.NewReader(updateBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestDeleteUser_OK(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodDelete, "/users/sample_username", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}
