package errors

import (
	"errors"
	"fmt"
	"strings"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Context string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(code ErrorCode, context string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: getMessageFromCause(cause),
		Context: context,
		Cause:   cause,
	}
}

func WrapWithMessage(code ErrorCode, context string, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Context: context,
		Cause:   cause,
	}
}

func StackTrace(err error) string {
	var sb strings.Builder
	indent := 0

	for err != nil {
		prefix := strings.Repeat("  ", indent)
		if indent > 0 {
			sb.WriteString(prefix + "-> ")
		}

		if appErr, ok := err.(*AppError); ok {
			if appErr.Context != "" {
				sb.WriteString(appErr.Context)
			} else {
				sb.WriteString(appErr.Message)
			}
			sb.WriteString("\n")
			err = appErr.Cause
		} else {
			sb.WriteString(err.Error())
			sb.WriteString("\n")
			break
		}
		indent++
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

func GetCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeInternal
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func getMessageFromCause(cause error) string {
	if cause == nil {
		return ""
	}
	if appErr, ok := cause.(*AppError); ok {
		return appErr.Message
	}
	return cause.Error()
}
