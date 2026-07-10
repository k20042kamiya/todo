package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"notification/domain/entity"
	"notification/domain/repository"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type NotificationUsecase interface {
	CheckAndSendNotifications(ctx context.Context) error
}

type notificationUsecase struct {
	notificationRepo repository.NotificationRepository
	todoRepo         repository.TodoRepository
	userRepo         repository.UserRepository
	emailSender      EmailSender
	frontendURL      string
}

func NewNotificationUsecase(
	notificationRepo repository.NotificationRepository,
	todoRepo repository.TodoRepository,
	userRepo repository.UserRepository,
	emailSender EmailSender,
	frontendURL string,
) NotificationUsecase {
	return &notificationUsecase{
		notificationRepo: notificationRepo,
		todoRepo:         todoRepo,
		userRepo:         userRepo,
		emailSender:      emailSender,
		frontendURL:      frontendURL,
	}
}

func (u *notificationUsecase) CheckAndSendNotifications(ctx context.Context) error {
	slog.InfoContext(ctx, "通知チェック開始")

	todos, err := u.todoRepo.FindUncompletedTodosWithDueDate(ctx)
	if err != nil {
		return fmt.Errorf("未完了Todo取得に失敗: %w", err)
	}
	slog.InfoContext(ctx, "通知対象Todo取得完了", "count", len(todos))

	now := time.Now()
	today := now.UTC().Truncate(24 * time.Hour)
	sentCount := 0
	notYetDueCount := 0
	for _, todo := range todos {
		if todo.DueDate == nil {
			continue
		}

		dueDay := todo.DueDate.UTC().Truncate(24 * time.Hour)
		daysUntilDue := int(dueDay.Sub(today).Hours() / 24)

		var notifType string
		switch {
		case daysUntilDue < 0:
			notifType = entity.NotificationTypeOverdue
		case daysUntilDue <= 3:
			notifType = entity.NotificationTypeApproaching
		default:
			notYetDueCount++
			continue
		}

		slog.InfoContext(ctx, "通知判定", "todo_id", todo.ID, "type", notifType, "days_until_due", daysUntilDue)

		sent, err := u.sendNotificationIfNeeded(ctx, todo, notifType)
		if err != nil {
			return fmt.Errorf("致命的エラーのため処理を中断: %w", err)
		}
		if sent {
			sentCount++
		}
	}

	slog.InfoContext(ctx, "通知チェック完了", "sent", sentCount, "not_yet_due", notYetDueCount)
	return nil
}

func (u *notificationUsecase) sendNotificationIfNeeded(ctx context.Context, todo *entity.Todo, notifType string) (bool, error) {
	// 重複確認: 既に本日送信済みならスキップ。DBエラーは異常系として即中断
	existing, err := u.notificationRepo.FindTodayByTodoID(ctx, todo.ID)
	if err != nil {
		return false, fmt.Errorf("重複チェックに失敗: todoID=%d, %w", todo.ID, err)
	}
	if existing != nil {
		return false, nil
	}

	// ユーザー取得: 存在しない場合のみwarnでスキップ。それ以外のエラーは即中断
	user, err := u.userRepo.FindByID(ctx, todo.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			slog.WarnContext(ctx, "通知対象ユーザーが存在しないためスキップ", "todo_id", todo.ID, "user_id", todo.UserID)
			return false, nil
		}
		return false, fmt.Errorf("ユーザー取得に失敗: todoID=%d, userID=%d, %w", todo.ID, todo.UserID, err)
	}

	// メール送信（先に送信してから記録する）
	subject, body := u.buildEmailContent(todo, notifType)
	if err := u.emailSender.Send(ctx, user.Email, subject, body); err != nil {
		if errors.Is(err, entity.ErrInvalidRecipient) {
			slog.WarnContext(ctx, "メール送信失敗（無効な宛先）", "todo_id", todo.ID, "to", user.Email, "error", err)
			return false, nil
		}
		return false, fmt.Errorf("SESサービスエラー: todoID=%d, to=%s, %w", todo.ID, user.Email, err)
	}

	// 通知レコード保存（メール送信成功後）
	notification := &entity.Notification{
		TodoID: todo.ID,
		UserID: todo.UserID,
		Type:   notifType,
	}
	// DBエラーは異常系として即中断。メールは送信済みのため翌日再送されうる（at-least-once）
	if err := u.notificationRepo.Create(ctx, notification); err != nil {
		return true, fmt.Errorf("通知レコード保存に失敗（メールは送信済み）: todoID=%d, %w", todo.ID, err)
	}

	slog.InfoContext(ctx, "通知送信完了", "todo_id", todo.ID, "type", notifType, "to", user.Email)
	return true, nil
}

func (u *notificationUsecase) buildEmailContent(todo *entity.Todo, notifType string) (subject, body string) {
	switch notifType {
	case entity.NotificationTypeApproaching:
		subject = fmt.Sprintf("【期日間近】%s", todo.Title)
		body = fmt.Sprintf("TODOの期日が近づいています。\n\nタイトル: %s\n期日: %s\n\nTODOを確認する: %s\n\n期日までに完了してください。",
			todo.Title, todo.DueDate.Format("2006-01-02"), u.frontendURL)
	case entity.NotificationTypeOverdue:
		subject = fmt.Sprintf("【期日超過】%s", todo.Title)
		body = fmt.Sprintf("TODOの期日が過ぎています。\n\nタイトル: %s\n期日: %s\n\nTODOを確認する: %s\n\n早急に対応してください。",
			todo.Title, todo.DueDate.Format("2006-01-02"), u.frontendURL)
	}
	return
}
