package errors

import "net/http"

type ErrorCode string

const (
	ErrCodeDatabase    ErrorCode = "DATABASE_ERROR"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrCodeValidation  ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden   ErrorCode = "FORBIDDEN"
	ErrCodeInternal    ErrorCode = "INTERNAL_ERROR"
)

func (c ErrorCode) HTTPStatus() int {
	switch c {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeValidation:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeDatabase, ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
