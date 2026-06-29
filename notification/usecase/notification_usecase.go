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
	userRepo         repository.UserRepository
	emailSender      EmailSender
}

func NewNotificationUsecase(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	emailSender EmailSender,
) NotificationUsecase {
	return &notificationUsecase{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		emailSender:      emailSender,
	}
}

func (u *notificationUsecase) CheckAndSendNotifications(ctx context.Context) error {
	todos, err := u.notificationRepo.FindUncompletedTodosWithDueDate(ctx)
	if err != nil {
		return fmt.Errorf("未完了Todo取得に失敗: %w", err)
	}

	now := time.Now()
	today := now.UTC().Truncate(24 * time.Hour)
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
			continue
		}

		if err := u.sendNotificationIfNeeded(ctx, todo, notifType); err != nil {
			return fmt.Errorf("致命的エラーのため処理を中断: %w", err)
		}
	}

	return nil
}

func (u *notificationUsecase) sendNotificationIfNeeded(ctx context.Context, todo *entity.Todo, notifType string) error {
	// 重複確認: DBエラー時は安全側（重複送信を避けるため）スキップ
	existing, err := u.notificationRepo.FindByTodoIDAndType(ctx, todo.ID, notifType)
	if err != nil {
		slog.WarnContext(ctx, "重複チェック失敗のためスキップ（安全側）", "todo_id", todo.ID, "error", err)
		return nil
	}
	if existing != nil {
		return nil
	}

	// ユーザー取得
	user, err := u.userRepo.FindByID(ctx, todo.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			slog.WarnContext(ctx, "通知対象ユーザーが存在しないためスキップ", "todo_id", todo.ID, "user_id", todo.UserID)
			return nil
		}
		slog.WarnContext(ctx, "ユーザー取得に失敗", "todo_id", todo.ID, "user_id", todo.UserID, "error", err)
		return nil
	}

	// メール送信（先に送信してから記録する）
	subject, body := u.buildEmailContent(todo, notifType)
	if err := u.emailSender.Send(ctx, user.Email, subject, body); err != nil {
		if errors.Is(err, entity.ErrInvalidRecipient) {
			slog.WarnContext(ctx, "メール送信失敗（無効な宛先）", "todo_id", todo.ID, "to", user.Email, "error", err)
			return nil
		}
		return fmt.Errorf("SESサービスエラー: todoID=%d, to=%s, %w", todo.ID, user.Email, err)
	}

	// 通知レコード保存（メール送信成功後）
	notification := &entity.Notification{
		TodoID: todo.ID,
		UserID: todo.UserID,
		Type:   notifType,
	}
	if err := u.notificationRepo.Create(ctx, notification); err != nil {
		slog.ErrorContext(ctx, "通知レコード保存失敗（メールは送信済み）", "todo_id", todo.ID, "error", err)
	}

	slog.InfoContext(ctx, "通知送信完了", "todo_id", todo.ID, "type", notifType, "to", user.Email)
	return nil
}

func (u *notificationUsecase) buildEmailContent(todo *entity.Todo, notifType string) (subject, body string) {
	switch notifType {
	case entity.NotificationTypeApproaching:
		subject = fmt.Sprintf("【期日間近】%s", todo.Title)
		body = fmt.Sprintf("TODOの期日が近づいています。\n\nタイトル: %s\n期日: %s\n\n期日までに完了してください。",
			todo.Title, todo.DueDate.Format("2006-01-02"))
	case entity.NotificationTypeOverdue:
		subject = fmt.Sprintf("【期日超過】%s", todo.Title)
		body = fmt.Sprintf("TODOの期日が過ぎています。\n\nタイトル: %s\n期日: %s\n\n早急に対応してください。",
			todo.Title, todo.DueDate.Format("2006-01-02"))
	}
	return
}
