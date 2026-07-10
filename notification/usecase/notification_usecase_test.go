package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"notification/domain/entity"
)

type mockNotificationRepository struct {
	findTodayByTodoIDFunc func(ctx context.Context, todoID int) (*entity.Notification, error)
	createFunc            func(ctx context.Context, notification *entity.Notification) error
}

func (m *mockNotificationRepository) FindTodayByTodoID(ctx context.Context, todoID int) (*entity.Notification, error) {
	return m.findTodayByTodoIDFunc(ctx, todoID)
}

func (m *mockNotificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	return m.createFunc(ctx, notification)
}

type mockTodoRepository struct {
	findUncompletedTodosWithDueDateFunc func(ctx context.Context) ([]*entity.Todo, error)
}

func (m *mockTodoRepository) FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error) {
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
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createdNotification = notification
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
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

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
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
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createdNotification = notification
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
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

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
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
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return &entity.Notification{ID: 1, TodoID: todoID, UserID: 1, Type: entity.NotificationTypeOverdue}, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
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
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
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
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createCalled {
		t.Error("期日が遠いTodoに対して通知が作成されてしまった")
	}
}

func TestCheckAndSendNotifications_UserNotFound(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 6, UserID: 99, Title: "ユーザー不在Todo", DueDate: &dueDate, IsCompleted: false},
	}

	createCalled := false
	sendCalled := false

	notifRepo := &mockNotificationRepository{
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, fmt.Errorf("user %d: %w", id, entity.ErrNotFound)
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			sendCalled = true
			return nil
		},
	}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("処理継続されるべきなのにエラーが返った: %v", err)
	}

	if sendCalled {
		t.Error("ユーザー不在なのにメールが送信された")
	}
	if createCalled {
		t.Error("ユーザー不在なのに通知レコードが作成された")
	}
}

func TestCheckAndSendNotifications_InvalidRecipient(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 7, UserID: 1, Title: "無効アドレスTodo", DueDate: &dueDate, IsCompleted: false},
	}

	createCalled := false

	notifRepo := &mockNotificationRepository{
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return &entity.User{ID: 1, Email: "invalid@@bad", Name: "Bad User"}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			return fmt.Errorf("ses rejected: %w", entity.ErrInvalidRecipient)
		},
	}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(context.Background()); err != nil {
		t.Fatalf("処理継続されるべきなのにエラーが返った: %v", err)
	}

	if createCalled {
		t.Error("無効アドレスエラーなのに通知レコードが作成された")
	}
}

func TestCheckAndSendNotifications_SESServiceError(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 8, UserID: 1, Title: "SES障害Todo", DueDate: &dueDate, IsCompleted: false},
	}

	notifRepo := &mockNotificationRepository{
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return &entity.User{ID: 1, Email: "test@example.com", Name: "Test User"}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			return errors.New("SES service unavailable")
		},
	}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(context.Background()); err == nil {
		t.Fatal("SESサービス障害時にエラーが返らなかった")
	}
}

func TestCheckAndSendNotifications_DuplicateCheckDBError(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 9, UserID: 1, Title: "重複確認DBエラーTodo", DueDate: &dueDate, IsCompleted: false},
	}

	createCalled := false
	sendCalled := false

	notifRepo := &mockNotificationRepository{
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, errors.New("db connection lost")
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			createCalled = true
			return nil
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return &entity.User{ID: 1, Email: "test@example.com", Name: "Test User"}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			sendCalled = true
			return nil
		},
	}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(context.Background()); err == nil {
		t.Fatal("重複確認DBエラーは異常系として中断されるべきなのにエラーが返らなかった")
	}

	if sendCalled {
		t.Error("重複確認DBエラー時にメールが送信された（中断されるべき）")
	}
	if createCalled {
		t.Error("重複確認DBエラー時に通知レコードが作成された（中断されるべき）")
	}
}

func TestCheckAndSendNotifications_ContextExpired(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 11, UserID: 1, Title: "タイムアウトTodo", DueDate: &dueDate, IsCompleted: false},
	}

	sendCalled := false

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // バッチタイムアウト発動後の状態を再現

	notifRepo := &mockNotificationRepository{
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, ctx.Err()
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{}
	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			sendCalled = true
			return nil
		},
	}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(ctx); err == nil {
		t.Fatal("コンテキスト期限切れ時にエラーが返らなかった（exit 1で検知されるべき）")
	}

	if sendCalled {
		t.Error("コンテキスト期限切れ後にメールが送信された")
	}
}

func TestCheckAndSendNotifications_RecordSaveFailure(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	todos := []*entity.Todo{
		{ID: 10, UserID: 1, Title: "レコード保存失敗Todo", DueDate: &dueDate, IsCompleted: false},
	}

	sendCalled := false

	notifRepo := &mockNotificationRepository{
		findTodayByTodoIDFunc: func(ctx context.Context, todoID int) (*entity.Notification, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, notification *entity.Notification) error {
			return errors.New("db write failed")
		},
	}

	todoRepo := &mockTodoRepository{
		findUncompletedTodosWithDueDateFunc: func(ctx context.Context) ([]*entity.Todo, error) {
			return todos, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return &entity.User{ID: 1, Email: "test@example.com", Name: "Test User"}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, to, subject, body string) error {
			sendCalled = true
			return nil
		},
	}

	uc := NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, "https://example.com")
	if err := uc.CheckAndSendNotifications(context.Background()); err == nil {
		t.Fatal("レコード保存失敗は異常系として中断されるべきなのにエラーが返らなかった")
	}

	if !sendCalled {
		t.Error("メールが送信されていない（レコード保存失敗前にメール送信されるべき）")
	}
}
