package todo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"todo/shared/appcontext"
	apperrors "todo/shared/errors"

	"github.com/labstack/echo/v4"
)

type mockUsecase struct {
	getTodosByUserIDFunc func(ctx context.Context, userID int) ([]*Todo, error)
	createTodoFunc       func(ctx context.Context, userID int, input CreateInput) (*Todo, error)
	updateTodoFunc       func(ctx context.Context, userID int, todoID int, input UpdateInput) (*Todo, error)
	deleteTodoFunc       func(ctx context.Context, userID int, todoID int) error
}

func (m *mockUsecase) GetTodosByUserID(ctx context.Context, userID int) ([]*Todo, error) {
	return m.getTodosByUserIDFunc(ctx, userID)
}

func (m *mockUsecase) CreateTodo(ctx context.Context, userID int, input CreateInput) (*Todo, error) {
	if m.createTodoFunc != nil {
		return m.createTodoFunc(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockUsecase) UpdateTodo(ctx context.Context, userID int, todoID int, input UpdateInput) (*Todo, error) {
	if m.updateTodoFunc != nil {
		return m.updateTodoFunc(ctx, userID, todoID, input)
	}
	return nil, nil
}

func (m *mockUsecase) DeleteTodo(ctx context.Context, userID int, todoID int) error {
	if m.deleteTodoFunc != nil {
		return m.deleteTodoFunc(ctx, userID, todoID)
	}
	return nil
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(&mockUsecase{})
	if h == nil {
		t.Error("NewHandler should return non-nil handler")
	}
}

func TestHandler_GetTodos(t *testing.T) {
	now := time.Now()
	content := "テスト内容"

	tests := []struct {
		name           string
		mockReturn     []*Todo
		mockError      error
		expectedStatus int
		expectedLen    int
	}{
		{
			name: "正常系: TODOリストを取得",
			mockReturn: []*Todo{
				{ID: 1, UserID: 1, Title: "テストTODO1", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now},
				{ID: 2, UserID: 1, Title: "テストTODO2", Content: nil, IsCompleted: true, CreatedAt: now, UpdatedAt: now},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:           "正常系: 空のTODOリスト",
			mockReturn:     []*Todo{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:           "異常系: ユースケースがエラーを返す",
			mockReturn:     nil,
			mockError:      apperrors.New(apperrors.ErrCodeDatabase, "database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/todos", nil)
			ctx := appcontext.SetUserID(req.Context(), 1)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			uc := &mockUsecase{
				getTodosByUserIDFunc: func(ctx context.Context, userID int) ([]*Todo, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			h := NewHandler(uc)
			err := h.GetTodos(c)

			if err != nil {
				t.Errorf("handler returned error: %v", err)
			}
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response TodoListResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if len(response.Todos) != tt.expectedLen {
					t.Errorf("expected %d todos, got %d", tt.expectedLen, len(response.Todos))
				}
				if tt.expectedLen > 0 && response.Todos[0].Title != tt.mockReturn[0].Title {
					t.Errorf("expected title %s, got %s", tt.mockReturn[0].Title, response.Todos[0].Title)
				}
			}
		})
	}
}

func TestHandler_GetTodos_Unauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(&mockUsecase{})
	err := h.GetTodos(c)

	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestHandler_CreateTodo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    string
		mockReturn     *Todo
		mockError      error
		expectedStatus int
	}{
		{
			name:        "正常系: TODOを作成",
			requestBody: `{"title":"テストTODO","content":"テスト内容"}`,
			mockReturn:  &Todo{ID: 1, UserID: 1, Title: "テストTODO", IsCompleted: false, CreatedAt: now, UpdatedAt: now},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "異常系: タイトルが空",
			requestBody:    `{"title":"","content":"テスト内容"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: 不正なJSON",
			requestBody:    `{"title":}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: タイトルが101文字",
			requestBody:    `{"title":"` + strings.Repeat("あ", 101) + `"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "正常系: タイトルがちょうど100文字",
			requestBody:    `{"title":"` + strings.Repeat("あ", 100) + `"}`,
			mockReturn:     &Todo{ID: 1, UserID: 1, Title: strings.Repeat("あ", 100), IsCompleted: false},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "異常系: ユースケースがエラーを返す",
			requestBody:    `{"title":"テストTODO"}`,
			mockError:      apperrors.New(apperrors.ErrCodeDatabase, "database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := appcontext.SetUserID(req.Context(), 1)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			uc := &mockUsecase{
				createTodoFunc: func(ctx context.Context, userID int, input CreateInput) (*Todo, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			h := NewHandler(uc)
			err := h.CreateTodo(c)

			if err != nil {
				t.Errorf("handler returned error: %v", err)
			}
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandler_UpdateTodo(t *testing.T) {
	now := time.Now()
	content := "更新後の内容"

	tests := []struct {
		name           string
		todoID         string
		requestBody    string
		mockReturn     *Todo
		mockError      error
		expectedStatus int
	}{
		{
			name:        "正常系: TODOを更新",
			todoID:      "1",
			requestBody: `{"title":"更新後のタイトル","content":"更新後の内容","is_completed":true}`,
			mockReturn:  &Todo{ID: 1, UserID: 1, Title: "更新後のタイトル", Content: &content, IsCompleted: true, CreatedAt: now, UpdatedAt: now},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "正常系: due_dateをYYYY-MM-DD形式で更新",
			todoID:      "1",
			requestBody: `{"title":"更新","due_date":"2026-06-30","is_completed":false}`,
			mockReturn:  &Todo{ID: 1, UserID: 1, Title: "更新", IsCompleted: false, CreatedAt: now, UpdatedAt: now},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "異常系: due_dateがISO 8601形式",
			todoID:         "1",
			requestBody:    `{"title":"更新","due_date":"2026-06-30T00:00:00Z","is_completed":false}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: タイトルが空",
			todoID:         "1",
			requestBody:    `{"title":"","content":"テスト内容"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: 不正なTODO ID",
			todoID:         "invalid",
			requestBody:    `{"title":"テストTODO"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: タイトルが101文字",
			todoID:         "1",
			requestBody:    `{"title":"` + strings.Repeat("あ", 101) + `"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: 権限がない",
			todoID:         "1",
			requestBody:    `{"title":"テストTODO"}`,
			mockError:      apperrors.New(apperrors.ErrCodeForbidden, "forbidden"),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "異常系: TODOが見つからない",
			todoID:         "999",
			requestBody:    `{"title":"テストTODO"}`,
			mockError:      apperrors.New(apperrors.ErrCodeNotFound, "todo not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "異常系: ユースケースがエラーを返す",
			todoID:         "1",
			requestBody:    `{"title":"テストTODO"}`,
			mockError:      apperrors.New(apperrors.ErrCodeDatabase, "database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/todos/"+tt.todoID, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := appcontext.SetUserID(req.Context(), 1)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.todoID)

			uc := &mockUsecase{
				updateTodoFunc: func(ctx context.Context, userID int, todoID int, input UpdateInput) (*Todo, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			h := NewHandler(uc)
			err := h.UpdateTodo(c)

			if err != nil {
				t.Errorf("handler returned error: %v", err)
			}
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response TodoResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.Title != tt.mockReturn.Title {
					t.Errorf("expected title %s, got %s", tt.mockReturn.Title, response.Title)
				}
			}
		})
	}
}

func TestHandler_DeleteTodo(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "正常系: TODOを削除",
			todoID:         "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "異常系: 不正なTODO ID",
			todoID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "異常系: 権限がない",
			todoID:         "1",
			mockError:      apperrors.New(apperrors.ErrCodeForbidden, "forbidden"),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "異常系: TODOが見つからない",
			todoID:         "999",
			mockError:      apperrors.New(apperrors.ErrCodeNotFound, "todo not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "異常系: ユースケースがエラーを返す",
			todoID:         "1",
			mockError:      apperrors.New(apperrors.ErrCodeDatabase, "database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/todos/"+tt.todoID, nil)
			ctx := appcontext.SetUserID(req.Context(), 1)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.todoID)

			uc := &mockUsecase{
				deleteTodoFunc: func(ctx context.Context, userID int, todoID int) error {
					return tt.mockError
				},
			}

			h := NewHandler(uc)
			err := h.DeleteTodo(c)

			if err != nil {
				t.Errorf("handler returned error: %v", err)
			}
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandler_DeleteTodo_Unauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewHandler(&mockUsecase{})
	err := h.DeleteTodo(c)

	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestDateOnly_MarshalJSON(t *testing.T) {
	d := DateOnly{time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if string(b) != `"2026-06-30"` {
		t.Errorf("expected %q, got %q", `"2026-06-30"`, string(b))
	}
}

func TestDateOnly_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantDay int
	}{
		{"正常系: YYYY-MM-DD形式", `"2026-06-30"`, false, 30},
		{"異常系: ISO 8601形式は不可", `"2026-06-30T00:00:00Z"`, true, 0},
		{"異常系: 不正な文字列", `"not-a-date"`, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d DateOnly
			err := json.Unmarshal([]byte(tt.input), &d)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && d.Day() != tt.wantDay {
				t.Errorf("expected day %d, got %d", tt.wantDay, d.Day())
			}
		})
	}
}

func TestTodoResponse_JSONFields(t *testing.T) {
	now := time.Now()
	content := "テスト内容"
	dueDate := DateOnly{now.Add(24 * time.Hour)}

	response := TodoResponse{
		ID: 1, UserID: 1, Title: "テストタイトル", Content: &content,
		DueDate: &dueDate, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	expectedFields := []string{"id", "user_id", "title", "content", "due_date", "is_completed", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("expected field %s not found in JSON", field)
		}
	}
}
