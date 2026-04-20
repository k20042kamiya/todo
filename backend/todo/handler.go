package todo

import (
	"net/http"
	"strconv"
	"time"

	"todo/shared/appcontext"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

type createTodoRequest struct {
	Title   string     `json:"title"`
	Content *string    `json:"content"`
	DueDate *time.Time `json:"due_date"`
}

type updateTodoRequest struct {
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	DueDate     *time.Time `json:"due_date"`
	IsCompleted bool       `json:"is_completed"`
}

type TodoResponse struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Title       string     `json:"title"`
	Content     *string    `json:"content"`
	DueDate     *time.Time `json:"due_date"`
	IsCompleted bool       `json:"is_completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoListResponse struct {
	Todos []*TodoResponse `json:"todos"`
}

func toResponse(todo *Todo) *TodoResponse {
	return &TodoResponse{
		ID:          todo.ID,
		UserID:      todo.UserID,
		Title:       todo.Title,
		Content:     todo.Content,
		DueDate:     todo.DueDate,
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get todos"})
	}

	response := &TodoListResponse{Todos: make([]*TodoResponse, len(todos))}
	for i, todo := range todos {
		response.Todos[i] = toResponse(todo)
	}

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

	todo, err := h.usecase.CreateTodo(ctx, userID, CreateInput{
		Title:   req.Title,
		Content: req.Content,
		DueDate: req.DueDate,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create todo"})
	}

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

	todo, err := h.usecase.UpdateTodo(ctx, userID, todoID, UpdateInput{
		Title:       req.Title,
		Content:     req.Content,
		DueDate:     req.DueDate,
		IsCompleted: req.IsCompleted,
	})
	if err != nil {
		if err.Error() == "forbidden" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You don't have permission to update this todo"})
		}
		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Todo not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update todo"})
	}

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
		if err.Error() == "forbidden" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You don't have permission to delete this todo"})
		}
		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Todo not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete todo"})
	}

	return c.NoContent(http.StatusNoContent)
}
