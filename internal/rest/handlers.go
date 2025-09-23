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

func (handler *Handler) UpdateUser(context *gin.Context) {
	username := context.Param("username")

	var updateRequest service.UpdateUserRequest
	if err := context.ShouldBindJSON(&updateRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser, err := handler.userService.Update(
		context.Request.Context(),
		username,
		updateRequest,
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, updatedUser)
}

func (handler *Handler) DeleteUser(context *gin.Context) {
	username := context.Param("username")
	requestContext := context.Request.Context()

	if deleteError := handler.userService.Delete(requestContext, username); deleteError != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": deleteError.Error()})
		return
	}

	context.Status(http.StatusNoContent)
}

func (handler *Handler) GetUser(context *gin.Context) {
	requestContext := context.Request.Context()
	username := context.Param("username")

	user, err := handler.userService.Get(requestContext, username)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, user)
}
