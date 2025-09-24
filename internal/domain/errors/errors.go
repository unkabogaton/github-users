package errors

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrorCodeValidation   ErrorCode = "validation_error"
	ErrorCodeNotFound     ErrorCode = "not_found"
	ErrorCodeConflict     ErrorCode = "conflict"
	ErrorCodeUnauthorized ErrorCode = "unauthorized"
	ErrorCodeForbidden    ErrorCode = "forbidden"
	ErrorCodeRateLimited  ErrorCode = "rate_limited"
	ErrorCodeUpstream     ErrorCode = "upstream_error"
	ErrorCodeInternal     ErrorCode = "internal_error"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *DomainError) Unwrap() error { return e.Cause }

func New(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func Wrap(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

func IsCode(err error, code ErrorCode) bool {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code == code
	}
	return false
}
