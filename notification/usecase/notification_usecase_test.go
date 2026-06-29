package usecase

import (
	"context"
	"testing"
	"time"

	"notification/domain/entity"
)

type mockNotificationRepository struct {
	findByTodoIDAndTypeFunc            func(ctx context.Context, todoID int, notifType string) (*entity.Notification, error)
	createFunc                         func(ctx context.Context, notification *entity.Notification) error
	findUncompletedTodosWithDueDateFunc func(ctx context.Context) ([]*entity.Todo, error)
}

func (m *mockNotificationRepository) FindByTodoIDAndType(ctx context.Context, todoID int, notifType string) (*entity.Notification, error) {
	return m.findByTodoIDAndTypeFunc(ctx, todoID, notifType)
}

func (m *mockNotificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	return m.createFunc(ctx, notification)
}

func (m *mockNotificationRepository) FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error) {
	return m.findUncompletedTodosWithDueDateFunc(ctx)
}

type mockUserRepository struct {
	findByIDFunc func(ctx context.Context, id int) (*entity.User, error)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*entity.User, error) {
	return m.findByIDFunc(ctx, id)
}

type mockEmailSender struct {
	sendFunc func(ctx context.Context, to, subject, body string) error
}

func (m *mockEmailSender) Send(ctx context.Context, to, subject, body string) error {
	return m.sendFunc(ctx, to, subject, body)
}

func TestCheckAndSendNotifications_Approaching(t *testing.T) {
	dueDate := time.Now().Add(48 * time.Hour)
	todos := []*entity.Todo{
		{ID: 1, UserID: 1, Title: "テストTodo", DueDate: &dueDate, IsCompleted: false},
	}

	var createdNotification *entity.Notification
	var sentEmail struct{ to, subject, body string }

	notifRepo := &mockNotificationRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
		findByTodoIDAndTypeFunc: func(ctx context.Context, todoID int, notifType string) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createdNotification = notification
			return nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return &entity.User{ID: 1, Email: "test@example.com", Name: "Test User"}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			sentEmail.to = to
			sentEmail.subject = subject
			sentEmail.body = body
			return nil
		},
	}

	uc := NewNotificationUsecase(notifRepo, userRepo, emailSender)
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createdNotification == nil {
		t.Fatal("通知レコードが作成されていない")
	}
	if createdNotification.Type != entity.NotificationTypeApproaching {
		t.Errorf("通知種別が期待と異なる: got=%s, want=%s", createdNotification.Type, entity.NotificationTypeApproaching)
	}
	if sentEmail.to != "test@example.com" {
		t.Errorf("送信先が期待と異なる: got=%s, want=test@example.com", sentEmail.to)
	}
}

func TestCheckAndSendNotifications_Overdue(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 2, UserID: 1, Title: "期限切れTodo", DueDate: &dueDate, IsCompleted: false},
	}

	var createdNotification *entity.Notification

	notifRepo := &mockNotificationRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
		findByTodoIDAndTypeFunc: func(ctx context.Context, todoID int, notifType string) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createdNotification = notification
			return nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return &entity.User{ID: 1, Email: "test@example.com", Name: "Test User"}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			return nil
		},
	}

	uc := NewNotificationUsecase(notifRepo, userRepo, emailSender)
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createdNotification == nil {
		t.Fatal("通知レコードが作成されていない")
	}
	if createdNotification.Type != entity.NotificationTypeOverdue {
		t.Errorf("通知種別が期待と異なる: got=%s, want=%s", createdNotification.Type, entity.NotificationTypeOverdue)
	}
}

func TestCheckAndSendNotifications_DuplicateSkip(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 3, UserID: 1, Title: "通知済みTodo", DueDate: &dueDate, IsCompleted: false},
	}

	createCalled := false

	notifRepo := &mockNotificationRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
		findByTodoIDAndTypeFunc: func(ctx context.Context, todoID int, notifType string) (*entity.Notification, error) {
			return &entity.Notification{ID: 1, TodoID: todoID, UserID: 1, Type: notifType}, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{}

	uc := NewNotificationUsecase(notifRepo, userRepo, emailSender)
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createCalled {
		t.Error("重複通知がスキップされていない")
	}
}

func TestCheckAndSendNotifications_NoDueDate(t *testing.T) {
	todos := []*entity.Todo{
		{ID: 4, UserID: 1, Title: "期日なしTodo", DueDate: nil, IsCompleted: false},
	}

	createCalled := false

	notifRepo := &mockNotificationRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{}

	uc := NewNotificationUsecase(notifRepo, userRepo, emailSender)
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createCalled {
		t.Error("期日なしTodoに対して通知が作成されてしまった")
	}
}

func TestCheckAndSendNotifications_FarFutureDueDate(t *testing.T) {
	dueDate := time.Now().Add(240 * time.Hour)
	todos := []*entity.Todo{
		{ID: 5, UserID: 1, Title: "まだ先のTodo", DueDate: &dueDate, IsCompleted: false},
	}

	createCalled := false

	notifRepo := &mockNotificationRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{}

	uc := NewNotificationUsecase(notifRepo, userRepo, emailSender)
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createCalled {
		t.Error("期日が遠いTodoに対して通知が作成されてしまった")
	}
}
