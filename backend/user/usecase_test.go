package user

import (
	"context"
	"errors"
	"testing"

	apperrors "todo/shared/errors"
)

type mockRepository struct {
	findByFirebaseUIDFunc func(ctx context.Context, firebaseUID string) (*User, error)
	findOrCreateFunc      func(ctx context.Context, user *User) error
}

func (m *mockRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error) {
	if m.findByFirebaseUIDFunc != nil {
		return m.findByFirebaseUIDFunc(ctx, firebaseUID)
	}
	return nil, apperrors.New(apperrors.ErrCodeNotFound, "user not found")
}

func (m *mockRepository) FindOrCreate(ctx context.Context, user *User) error {
	if m.findOrCreateFunc != nil {
		return m.findOrCreateFunc(ctx, user)
	}
	return nil
}

func TestUsecase_FindOrCreateByFirebaseUID(t *testing.T) {
	existingUser := &User{ID: 1, FirebaseUID: "uid-abc", Email: "test@example.com", Name: "テストユーザー"}

	tests := []struct {
		name                  string
		firebaseUID           string
		email                 string
		inputName             string
		findByFirebaseUIDFunc func(ctx context.Context, firebaseUID string) (*User, error)
		findOrCreateFunc      func(ctx context.Context, user *User) error
		expectedID            int
		expectedName          string
		expectedError         bool
	}{
		{
			name:        "正常系: 既存ユーザーが見つかる",
			firebaseUID: "uid-abc",
			email:       "test@example.com",
			inputName:   "テストユーザー",
			findByFirebaseUIDFunc: func(ctx context.Context, firebaseUID string) (*User, error) {
				return existingUser, nil
			},
			expectedID:    1,
			expectedName:  "テストユーザー",
			expectedError: false,
		},
		{
			name:        "正常系: 新規ユーザーを作成",
			firebaseUID: "uid-new",
			email:       "new@example.com",
			inputName:   "新規ユーザー",
			findByFirebaseUIDFunc: func(ctx context.Context, firebaseUID string) (*User, error) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "user not found")
			},
			findOrCreateFunc: func(ctx context.Context, user *User) error {
				user.ID = 2
				return nil
			},
			expectedID:    2,
			expectedName:  "新規ユーザー",
			expectedError: false,
		},
		{
			name:        "正常系: nameが空の場合はUnknownが設定される",
			firebaseUID: "uid-noname",
			email:       "noname@example.com",
			inputName:   "",
			findByFirebaseUIDFunc: func(ctx context.Context, firebaseUID string) (*User, error) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "user not found")
			},
			findOrCreateFunc: func(ctx context.Context, user *User) error {
				user.ID = 3
				return nil
			},
			expectedID:    3,
			expectedName:  "Unknown",
			expectedError: false,
		},
		{
			name:        "正常系: 競合時に既存ユーザーを返す（INSERT IGNOREのシミュレート）",
			firebaseUID: "uid-race",
			email:       "race@example.com",
			inputName:   "競合ユーザー",
			findByFirebaseUIDFunc: func(ctx context.Context, firebaseUID string) (*User, error) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "user not found")
			},
			findOrCreateFunc: func(ctx context.Context, user *User) error {
				// 競合してINSERT IGNOREがスキップされた後、既存レコードを取得するケース
				user.ID = 4
				user.Name = "競合ユーザー"
				return nil
			},
			expectedID:    4,
			expectedName:  "競合ユーザー",
			expectedError: false,
		},
		{
			name:        "異常系: DB障害（Not Found以外のエラー）",
			firebaseUID: "uid-dberr",
			email:       "err@example.com",
			inputName:   "エラーユーザー",
			findByFirebaseUIDFunc: func(ctx context.Context, firebaseUID string) (*User, error) {
				return nil, apperrors.Wrap(apperrors.ErrCodeDatabase, "FindByFirebaseUID", errors.New("connection refused"))
			},
			expectedError: true,
		},
		{
			name:        "異常系: FindOrCreateがエラーを返す",
			firebaseUID: "uid-createerr",
			email:       "createerr@example.com",
			inputName:   "作成エラーユーザー",
			findByFirebaseUIDFunc: func(ctx context.Context, firebaseUID string) (*User, error) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "user not found")
			},
			findOrCreateFunc: func(ctx context.Context, user *User) error {
				return apperrors.Wrap(apperrors.ErrCodeDatabase, "FindOrCreate user", errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				findByFirebaseUIDFunc: tt.findByFirebaseUIDFunc,
				findOrCreateFunc:      tt.findOrCreateFunc,
			}

			uc := NewUsecase(repo)
			user, err := uc.FindOrCreateByFirebaseUID(context.Background(), tt.firebaseUID, tt.email, tt.inputName)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if user.ID != tt.expectedID {
				t.Errorf("expected ID %d, got %d", tt.expectedID, user.ID)
			}
			if user.Name != tt.expectedName {
				t.Errorf("expected Name %q, got %q", tt.expectedName, user.Name)
			}
		})
	}
}
