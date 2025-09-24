package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    derr "github.com/unkabogaton/github-users/internal/domain/errors"
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
        _ = ctx.Error(listErr)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) GetUser(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	username := ctx.Param("username")

	user, err := c.userService.Get(requestContext, username)
	if err != nil {
        // If service did not categorize, treat not found sensibly at controller level
        if derr.IsCode(err, derr.ErrorCodeNotFound) {
            _ = ctx.Error(err)
        } else {
            _ = ctx.Error(derr.Wrap(derr.ErrorCodeInternal, "failed to get user", err))
        }
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	username := ctx.Param("username")

	var updateRequest interfaces.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&updateRequest); err != nil {
        _ = ctx.Error(derr.Wrap(derr.ErrorCodeValidation, "invalid request payload", err))
		return
	}

	updatedUser, err := c.userService.Update(
		ctx.Request.Context(),
		username,
		updateRequest,
	)
	if err != nil {
        _ = ctx.Error(derr.Wrap(derr.ErrorCodeInternal, "failed to update user", err))
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	username := ctx.Param("username")
	requestContext := ctx.Request.Context()

	if deleteError := c.userService.Delete(requestContext, username); deleteError != nil {
        _ = ctx.Error(derr.Wrap(derr.ErrorCodeInternal, "failed to delete user", deleteError))
		return
	}

	ctx.Status(http.StatusNoContent)
}
