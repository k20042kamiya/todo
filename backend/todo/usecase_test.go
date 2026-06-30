package todo

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockTransactionManager struct{}

func (m *mockTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type mockRepository struct {
	findByUserIDFunc func(ctx context.Context, userID int) ([]*Todo, error)
	findByIDFunc     func(ctx context.Context, id int) (*Todo, error)
	createFunc       func(ctx context.Context, todo *Todo) error
	updateFunc       func(ctx context.Context, todo *Todo) error
	deleteFunc       func(ctx context.Context, id int, userID int) error
}

func (m *mockRepository) FindByUserID(ctx context.Context, userID int) ([]*Todo, error) {
	return m.findByUserIDFunc(ctx, userID)
}

func (m *mockRepository) FindByID(ctx context.Context, id int) (*Todo, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) Create(ctx context.Context, todo *Todo) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, todo)
	}
	return nil
}

func (m *mockRepository) Update(ctx context.Context, todo *Todo) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, todo)
	}
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id int, userID int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, userID)
	}
	return nil
}

func TestNewUsecase(t *testing.T) {
	uc := NewUsecase(&mockTransactionManager{}, &mockRepository{})
	if uc == nil {
		t.Error("NewUsecase should return non-nil usecase")
	}
}

func TestUsecase_GetTodosByUserID(t *testing.T) {
	now := time.Now()
	content := "テスト内容"

	tests := []struct {
		name          string
		userID        int
		mockReturn    []*Todo
		mockError     error
		expectedLen   int
		expectedError bool
	}{
		{
			name:   "正常系: TODOが取得できる",
			userID: 1,
			mockReturn: []*Todo{
				{ID: 1, UserID: 1, Title: "テストTODO1", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now},
				{ID: 2, UserID: 1, Title: "テストTODO2", Content: nil, IsCompleted: true, CreatedAt: now, UpdatedAt: now},
			},
			mockError:     nil,
			expectedLen:   2,
			expectedError: false,
		},
		{
			name:          "正常系: TODOが0件の場合",
			userID:        2,
			mockReturn:    []*Todo{},
			mockError:     nil,
			expectedLen:   0,
			expectedError: false,
		},
		{
			name:          "異常系: リポジトリがエラーを返す",
			userID:        3,
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedLen:   0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				findByUserIDFunc: func(ctx context.Context, userID int) ([]*Todo, error) {
					if userID != tt.userID {
						t.Errorf("expected userID %d, got %d", tt.userID, userID)
					}
					return tt.mockReturn, tt.mockError
				},
			}

			uc := NewUsecase(&mockTransactionManager{}, repo)
			todos, err := uc.GetTodosByUserID(context.Background(), tt.userID)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(todos) != tt.expectedLen {
					t.Errorf("expected %d todos, got %d", tt.expectedLen, len(todos))
				}
			}
		})
	}
}

func TestUsecase_GetTodosByUserID_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &mockRepository{
		findByUserIDFunc: func(ctx context.Context, userID int) ([]*Todo, error) {
			return nil, ctx.Err()
		},
	}

	uc := NewUsecase(&mockTransactionManager{}, repo)
	_, err := uc.GetTodosByUserID(ctx, 1)

	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

func TestUsecase_UpdateTodo(t *testing.T) {
	now := time.Now()
	content := "テスト内容"
	updatedContent := "更新後の内容"

	tests := []struct {
		name           string
		userID         int
		todoID         int
		input          UpdateInput
		existingTodo   *Todo
		findByIDError  error
		updateError    error
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:   "正常系: TODOを更新",
			userID: 1,
			todoID: 1,
			input:  UpdateInput{Title: "更新後のタイトル", Content: &updatedContent, IsCompleted: true},
			existingTodo: &Todo{
				ID: 1, UserID: 1, Title: "元のタイトル", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
			},
			expectedError: false,
		},
		{
			name:           "異常系: TODOが見つからない",
			userID:         1,
			todoID:         999,
			input:          UpdateInput{Title: "更新後のタイトル"},
			findByIDError:  errors.New("record not found"),
			expectedError:  true,
			expectedErrMsg: "record not found",
		},
		{
			name:   "異常系: 他のユーザーのTODO",
			userID: 2,
			todoID: 1,
			input:  UpdateInput{Title: "更新後のタイトル"},
			existingTodo: &Todo{
				ID: 1, UserID: 1, Title: "元のタイトル", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
			},
			expectedError:  true,
			expectedErrMsg: "[FORBIDDEN] forbidden",
		},
		{
			name:   "異常系: 更新時にエラー",
			userID: 1,
			todoID: 1,
			input:  UpdateInput{Title: "更新後のタイトル"},
			existingTodo: &Todo{
				ID: 1, UserID: 1, Title: "元のタイトル", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
			},
			updateError:    errors.New("database error"),
			expectedError:  true,
			expectedErrMsg: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				findByIDFunc: func(ctx context.Context, id int) (*Todo, error) {
					if id != tt.todoID {
						t.Errorf("expected todoID %d, got %d", tt.todoID, id)
					}
					return tt.existingTodo, tt.findByIDError
				},
				updateFunc: func(ctx context.Context, todo *Todo) error {
					return tt.updateError
				},
			}

			uc := NewUsecase(&mockTransactionManager{}, repo)
			result, err := uc.UpdateTodo(context.Background(), tt.userID, tt.todoID, tt.input)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.expectedErrMsg != "" && err.Error() != tt.expectedErrMsg {
					t.Errorf("expected error message %s, got %s", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected result, got nil")
				}
				if result != nil && result.Title != tt.input.Title {
					t.Errorf("expected title %s, got %s", tt.input.Title, result.Title)
				}
				if result != nil && result.IsCompleted != tt.input.IsCompleted {
					t.Errorf("expected is_completed %v, got %v", tt.input.IsCompleted, result.IsCompleted)
				}
			}
		})
	}
}

func TestUsecase_DeleteTodo(t *testing.T) {
	now := time.Now()
	content := "テスト内容"

	tests := []struct {
		name           string
		userID         int
		todoID         int
		existingTodo   *Todo
		findByIDError  error
		deleteError    error
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:   "正常系: TODOを削除",
			userID: 1,
			todoID: 1,
			existingTodo: &Todo{
				ID: 1, UserID: 1, Title: "テストTODO", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
			},
			expectedError: false,
		},
		{
			name:           "異常系: TODOが見つからない",
			userID:         1,
			todoID:         999,
			findByIDError:  errors.New("record not found"),
			expectedError:  true,
			expectedErrMsg: "record not found",
		},
		{
			name:   "異常系: 他のユーザーのTODO",
			userID: 2,
			todoID: 1,
			existingTodo: &Todo{
				ID: 1, UserID: 1, Title: "テストTODO", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
			},
			expectedError:  true,
			expectedErrMsg: "[FORBIDDEN] forbidden",
		},
		{
			name:   "異常系: 削除時にエラー",
			userID: 1,
			todoID: 1,
			existingTodo: &Todo{
				ID: 1, UserID: 1, Title: "テストTODO", Content: &content, IsCompleted: false, CreatedAt: now, UpdatedAt: now,
			},
			deleteError:    errors.New("database error"),
			expectedError:  true,
			expectedErrMsg: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				findByIDFunc: func(ctx context.Context, id int) (*Todo, error) {
					if id != tt.todoID {
						t.Errorf("expected todoID %d, got %d", tt.todoID, id)
					}
					return tt.existingTodo, tt.findByIDError
				},
				deleteFunc: func(ctx context.Context, id int, userID int) error {
					return tt.deleteError
				},
			}

			uc := NewUsecase(&mockTransactionManager{}, repo)
			err := uc.DeleteTodo(context.Background(), tt.userID, tt.todoID)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.expectedErrMsg != "" && err.Error() != tt.expectedErrMsg {
					t.Errorf("expected error message %s, got %s", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
