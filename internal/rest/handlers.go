package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unkabogaton/github-users/internal/service"
)

type Handler struct {
	userService *service.UserService
}

func NewHandler(userService *service.UserService) *Handler {
	return &Handler{userService: userService}
}

func (handler *Handler) ListUsers(context *gin.Context) {
	requestContext := context.Request.Context()

	users, listErr := handler.userService.List(requestContext)
	if listErr != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": listErr.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, users)
}
