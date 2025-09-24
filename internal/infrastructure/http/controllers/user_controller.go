package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/unkabogaton/github-users/internal/domain/errors"
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

	limitQueryValue := ginContext.DefaultQuery("limit", "")
	if limitQueryValue != "" {
		if parsedLimit, parseError := strconv.Atoi(limitQueryValue); parseError == nil {
			defaultLimit = parsedLimit
		}
	}

	pageQueryValue := ginContext.DefaultQuery("page", "")
	if pageQueryValue != "" {
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

	userList, userListError := controller.userService.List(httpRequestContext, listOptions)
	if userListError != nil {
		_ = ginContext.Error(userListError)
		return
	}

	ginContext.JSON(http.StatusOK, userList)
}

func (controller *UserController) GetUser(ginContext *gin.Context) {
	httpRequestContext := ginContext.Request.Context()
	usernameParameter := ginContext.Param("username")

	userEntity, getUserError := controller.userService.Get(httpRequestContext, usernameParameter)
	if getUserError != nil {
		if domainErrors.IsCode(getUserError, domainErrors.ErrorCodeNotFound) {
			_ = ginContext.Error(getUserError)
		} else {
			_ = ginContext.Error(domainErrors.Wrap(domainErrors.ErrorCodeInternal, "failed to get user", getUserError))
		}
		return
	}

	ginContext.JSON(http.StatusOK, userEntity)
}

func (controller *UserController) UpdateUser(ginContext *gin.Context) {
	usernameParameter := ginContext.Param("username")

	var updateUserRequest interfaces.UpdateUserRequest
	if bindError := ginContext.ShouldBindJSON(&updateUserRequest); bindError != nil {
		_ = ginContext.Error(domainErrors.Wrap(domainErrors.ErrorCodeValidation, "invalid request payload", bindError))
		return
	}

	updatedUserEntity, updateError := controller.userService.Update(
		ginContext.Request.Context(),
		usernameParameter,
		updateUserRequest,
	)
	if updateError != nil {
		_ = ginContext.Error(domainErrors.Wrap(domainErrors.ErrorCodeInternal, "failed to update user", updateError))
		return
	}

	ginContext.JSON(http.StatusOK, updatedUserEntity)
}

func (controller *UserController) DeleteUser(ginContext *gin.Context) {
	usernameParameter := ginContext.Param("username")
	httpRequestContext := ginContext.Request.Context()

	deleteError := controller.userService.Delete(httpRequestContext, usernameParameter)
	if deleteError != nil {
		_ = ginContext.Error(domainErrors.Wrap(domainErrors.ErrorCodeInternal, "failed to delete user", deleteError))
		return
	}

	ginContext.JSON(http.StatusOK, gin.H{"username": usernameParameter})
}
