package middleware

import (
	stdErrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	derr "github.com/unkabogaton/github-users/internal/domain/errors"
	"github.com/unkabogaton/github-users/internal/infrastructure/http/models"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors.Last().Err

		statusCode, response := mapErrorToResponse(err)
		ctx.JSON(statusCode, response)
	}
}

func mapErrorToResponse(err error) (int, models.ErrorResponse) {
	var domainError *derr.DomainError
	if stdErrors.As(err, &domainError) {
		switch domainError.Code {
		case derr.ErrorCodeValidation:
			return http.StatusBadRequest, models.ErrorResponse{Code: http.StatusBadRequest, Message: domainError.Message}
		case derr.ErrorCodeNotFound:
			return http.StatusNotFound, models.ErrorResponse{Code: http.StatusNotFound, Message: domainError.Message}
		case derr.ErrorCodeConflict:
			return http.StatusConflict, models.ErrorResponse{Code: http.StatusConflict, Message: domainError.Message}
		case derr.ErrorCodeUnauthorized:
			return http.StatusUnauthorized, models.ErrorResponse{Code: http.StatusUnauthorized, Message: domainError.Message}
		case derr.ErrorCodeForbidden:
			return http.StatusForbidden, models.ErrorResponse{Code: http.StatusForbidden, Message: domainError.Message}
		case derr.ErrorCodeRateLimited:
			return http.StatusTooManyRequests, models.ErrorResponse{Code: http.StatusTooManyRequests, Message: domainError.Message}
		case derr.ErrorCodeUpstream:
			fallthrough
		case derr.ErrorCodeInternal:
			fallthrough
		default:
			return http.StatusInternalServerError, models.ErrorResponse{Code: http.StatusInternalServerError, Message: "An internal error occurred"}
		}
	}
	return http.StatusInternalServerError, models.ErrorResponse{Code: http.StatusInternalServerError, Message: "An internal error occurred"}
}
