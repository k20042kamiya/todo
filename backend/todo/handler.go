package todo

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"todo/shared/appcontext"
	apperrors "todo/shared/errors"

	"github.com/labstack/echo/v4"
)

type DateOnly struct{ time.Time }

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

func dateOnlyToTime(d *DateOnly) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

type createTodoRequest struct {
	Title   string    `json:"title"`
	Content *string   `json:"content"`
	DueDate *DateOnly `json:"due_date"`
}

type updateTodoRequest struct {
	Title       string    `json:"title"`
	Content     *string   `json:"content"`
	DueDate     *DateOnly `json:"due_date"`
	IsCompleted bool      `json:"is_completed"`
}

type TodoResponse struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Title       string     `json:"title"`
	Content     *string    `json:"content"`
	DueDate     *DateOnly  `json:"due_date"`
	IsCompleted bool       `json:"is_completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoListResponse struct {
	Todos []*TodoResponse `json:"todos"`
}

func toResponse(todo *Todo) *TodoResponse {
	var dueDate *DateOnly
	if todo.DueDate != nil {
		d := DateOnly{*todo.DueDate}
		dueDate = &d
	}
	return &TodoResponse{
		ID:          todo.ID,
		UserID:      todo.UserID,
		Title:       todo.Title,
		Content:     todo.Content,
		DueDate:     dueDate,
		IsCompleted: todo.IsCompleted,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}

func (h *Handler) GetTodos(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := appcontext.GetUserID(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	todos, err := h.usecase.GetTodosByUserID(ctx, userID)
	if err != nil {
		code := apperrors.GetCode(err)
		if code.HTTPStatus() >= 500 {
			slog.ErrorContext(ctx, "GetTodosByUserID failed", "userID", userID, "error", err)
		}
		return c.JSON(code.HTTPStatus(), map[string]string{"error": safeMessage(code, err)})
	}

	response := &TodoListResponse{Todos: make([]*TodoResponse, len(todos))}
	for i, todo := range todos {
		response.Todos[i] = toResponse(todo)
	}

	slog.InfoContext(ctx, "GetTodos succeeded", "userID", userID, "count", len(todos))

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateTodo(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := appcontext.GetUserID(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	var req createTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title is required"})
	}
	if len([]rune(req.Title)) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title must be 100 characters or less"})
	}

	todo, err := h.usecase.CreateTodo(ctx, userID, CreateInput{
		Title:   req.Title,
		Content: req.Content,
		DueDate: dateOnlyToTime(req.DueDate),
	})
	if err != nil {
		code := apperrors.GetCode(err)
		if code.HTTPStatus() >= 500 {
			slog.ErrorContext(ctx, "CreateTodo failed", "userID", userID, "error", err)
		}
		return c.JSON(code.HTTPStatus(), map[string]string{"error": safeMessage(code, err)})
	}

	slog.InfoContext(ctx, "CreateTodo succeeded", "userID", userID, "todoID", todo.ID)

	return c.JSON(http.StatusCreated, toResponse(todo))
}

func (h *Handler) UpdateTodo(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := appcontext.GetUserID(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid todo ID"})
	}

	var req updateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title is required"})
	}
	if len([]rune(req.Title)) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title must be 100 characters or less"})
	}

	todo, err := h.usecase.UpdateTodo(ctx, userID, todoID, UpdateInput{
		Title:       req.Title,
		Content:     req.Content,
		DueDate:     dateOnlyToTime(req.DueDate),
		IsCompleted: req.IsCompleted,
	})
	if err != nil {
		code := apperrors.GetCode(err)
		if code.HTTPStatus() >= 500 {
			slog.ErrorContext(ctx, "UpdateTodo failed", "userID", userID, "todoID", todoID, "error", err)
		}
		return c.JSON(code.HTTPStatus(), map[string]string{"error": safeMessage(code, err)})
	}

	slog.InfoContext(ctx, "UpdateTodo succeeded", "userID", userID, "todoID", todoID)

	return c.JSON(http.StatusOK, toResponse(todo))
}

func (h *Handler) DeleteTodo(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := appcontext.GetUserID(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid todo ID"})
	}

	err = h.usecase.DeleteTodo(ctx, userID, todoID)
	if err != nil {
		code := apperrors.GetCode(err)
		if code.HTTPStatus() >= 500 {
			slog.ErrorContext(ctx, "DeleteTodo failed", "userID", userID, "todoID", todoID, "error", err)
		}
		return c.JSON(code.HTTPStatus(), map[string]string{"error": safeMessage(code, err)})
	}

	slog.InfoContext(ctx, "DeleteTodo succeeded", "userID", userID, "todoID", todoID)

	return c.NoContent(http.StatusNoContent)
}

func safeMessage(code apperrors.ErrorCode, err error) string {
	switch code {
	case apperrors.ErrCodeNotFound, apperrors.ErrCodeForbidden,
		apperrors.ErrCodeValidation, apperrors.ErrCodeUnauthorized:
		return err.Error()
	default:
		return "Internal server error"
	}
}
