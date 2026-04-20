package errors

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "Causeなしの場合",
			appErr: &AppError{
				Code:    ErrCodeNotFound,
				Message: "リソースが見つかりません",
			},
			expected: "[NOT_FOUND] リソースが見つかりません",
		},
		{
			name: "Causeありの場合",
			appErr: &AppError{
				Code:    ErrCodeDatabase,
				Message: "データベースエラー",
				Cause:   errors.New("connection refused"),
			},
			expected: "[DATABASE_ERROR] データベースエラー: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appErr.Error()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	appErr := &AppError{
		Code:    ErrCodeInternal,
		Message: "内部エラー",
		Cause:   cause,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("expected %v, got %v", cause, unwrapped)
	}
}

func TestNew(t *testing.T) {
	appErr := New(ErrCodeValidation, "バリデーションエラー")

	if appErr.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, appErr.Code)
	}
	if appErr.Message != "バリデーションエラー" {
		t.Errorf("expected message %s, got %s", "バリデーションエラー", appErr.Message)
	}
	if appErr.Cause != nil {
		t.Errorf("expected nil cause, got %v", appErr.Cause)
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("original error")
	appErr := Wrap(ErrCodeDatabase, "repository.FindByID", cause)

	if appErr.Code != ErrCodeDatabase {
		t.Errorf("expected code %s, got %s", ErrCodeDatabase, appErr.Code)
	}
	if appErr.Context != "repository.FindByID" {
		t.Errorf("expected context %s, got %s", "repository.FindByID", appErr.Context)
	}
	if appErr.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, appErr.Cause)
	}
	if appErr.Message != "original error" {
		t.Errorf("expected message %s, got %s", "original error", appErr.Message)
	}
}

func TestWrap_WithAppErrorCause(t *testing.T) {
	innerErr := New(ErrCodeNotFound, "ユーザーが見つかりません")
	appErr := Wrap(ErrCodeInternal, "usecase.GetUser", innerErr)

	if appErr.Message != "ユーザーが見つかりません" {
		t.Errorf("expected message from inner AppError, got %s", appErr.Message)
	}
}

func TestWrapWithMessage(t *testing.T) {
	cause := errors.New("connection timeout")
	appErr := WrapWithMessage(ErrCodeDatabase, "repository.Save", "保存に失敗しました", cause)

	if appErr.Code != ErrCodeDatabase {
		t.Errorf("expected code %s, got %s", ErrCodeDatabase, appErr.Code)
	}
	if appErr.Message != "保存に失敗しました" {
		t.Errorf("expected message %s, got %s", "保存に失敗しました", appErr.Message)
	}
	if appErr.Context != "repository.Save" {
		t.Errorf("expected context %s, got %s", "repository.Save", appErr.Context)
	}
	if appErr.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, appErr.Cause)
	}
}

func TestStackTrace(t *testing.T) {
	innerErr := errors.New("SQL syntax error")
	midErr := Wrap(ErrCodeDatabase, "repository.Query", innerErr)
	outerErr := Wrap(ErrCodeInternal, "usecase.Execute", midErr)

	trace := StackTrace(outerErr)

	if !strings.Contains(trace, "usecase.Execute") {
		t.Errorf("trace should contain 'usecase.Execute', got %s", trace)
	}
	if !strings.Contains(trace, "repository.Query") {
		t.Errorf("trace should contain 'repository.Query', got %s", trace)
	}
	if !strings.Contains(trace, "SQL syntax error") {
		t.Errorf("trace should contain 'SQL syntax error', got %s", trace)
	}
}

func TestStackTrace_WithNonAppError(t *testing.T) {
	stdErr := errors.New("standard error")
	trace := StackTrace(stdErr)

	if trace != "standard error" {
		t.Errorf("expected 'standard error', got %s", trace)
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "AppErrorの場合",
			err:      New(ErrCodeNotFound, "not found"),
			expected: ErrCodeNotFound,
		},
		{
			name:     "ラップされたAppErrorの場合",
			err:      Wrap(ErrCodeDatabase, "context", New(ErrCodeValidation, "validation")),
			expected: ErrCodeDatabase,
		},
		{
			name:     "標準エラーの場合",
			err:      errors.New("standard error"),
			expected: ErrCodeInternal,
		},
		{
			name:     "nilの場合",
			err:      nil,
			expected: ErrCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCode(tt.err)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIs(t *testing.T) {
	target := errors.New("target error")
	wrapped := Wrap(ErrCodeInternal, "context", target)

	if !Is(wrapped, target) {
		t.Error("Is should return true for wrapped error")
	}
}

func TestAs(t *testing.T) {
	appErr := New(ErrCodeNotFound, "not found")
	wrapped := Wrap(ErrCodeInternal, "context", appErr)

	var target *AppError
	if !As(wrapped, &target) {
		t.Error("As should return true for AppError")
	}
	if target.Code != ErrCodeInternal {
		t.Errorf("expected code %s, got %s", ErrCodeInternal, target.Code)
	}
}

func TestErrorCode_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeValidation, http.StatusBadRequest},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeDatabase, http.StatusInternalServerError},
		{ErrCodeInternal, http.StatusInternalServerError},
		{ErrorCode("UNKNOWN"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			result := tt.code.HTTPStatus()
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetMessageFromCause_NilCause(t *testing.T) {
	appErr := Wrap(ErrCodeInternal, "context", nil)
	if appErr.Message != "" {
		t.Errorf("expected empty message, got %s", appErr.Message)
	}
}
