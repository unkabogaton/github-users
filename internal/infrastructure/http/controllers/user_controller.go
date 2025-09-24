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

func (controller *UserController) ListUsers(ginContext *gin.Context) {
    httpRequestContext := ginContext.Request.Context()

    defaultLimit := 10
    defaultPage := 1
    sortColumn := ginContext.DefaultQuery("orderby", "id")
    sortDirection := ginContext.DefaultQuery("order", "asc")

    if limitQueryValue := ginContext.DefaultQuery("limit", ""); limitQueryValue != "" {
        if parsedLimit, parseError := strconv.Atoi(limitQueryValue); parseError == nil {
            defaultLimit = parsedLimit
        }
    }

    if pageQueryValue := ginContext.DefaultQuery("page", ""); pageQueryValue != "" {
        if parsedPage, parseError := strconv.Atoi(pageQueryValue); parseError == nil {
            defaultPage = parsedPage
        }
    }

    listOptions := interfaces.ListOptions{
        Limit:          defaultLimit,
        Page:           defaultPage,
        OrderBy:        sortColumn,
        OrderDirection: sortDirection,
    }

    users, userListError := controller.userService.List(httpRequestContext, listOptions)
    if userListError != nil {
        _ = ginContext.Error(userListError)
        return
    }

    ginContext.JSON(http.StatusOK, users)
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

    ctx.JSON(http.StatusOK, gin.H{"username": username})
}
