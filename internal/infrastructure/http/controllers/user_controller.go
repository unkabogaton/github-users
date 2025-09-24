package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"
)

type UserController struct {
	userService interfaces.UserService
}

func NewUserController(userService interfaces.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) ListUsers(ctx *gin.Context) {
	requestContext := ctx.Request.Context()

	users, listErr := c.userService.List(requestContext)
	if listErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": listErr.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) GetUser(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	username := ctx.Param("username")

	user, err := c.userService.Get(requestContext, username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	username := ctx.Param("username")

	var updateRequest interfaces.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&updateRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser, err := c.userService.Update(
		ctx.Request.Context(),
		username,
		updateRequest,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	username := ctx.Param("username")
	requestContext := ctx.Request.Context()

	if deleteError := c.userService.Delete(requestContext, username); deleteError != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": deleteError.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
