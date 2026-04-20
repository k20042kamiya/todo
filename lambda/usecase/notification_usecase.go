package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"lambda/domain/entity"
	"lambda/domain/repository"
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
	for _, todo := range todos {
		if todo.DueDate == nil {
			continue
		}

		dueDate := *todo.DueDate
		daysUntilDue := time.Until(dueDate).Hours() / 24

		if daysUntilDue < 0 {
			if err := u.sendNotificationIfNeeded(ctx, todo, entity.NotificationTypeOverdue, now); err != nil {
				log.Printf("[WARN] overdue通知送信に失敗: todoID=%d, error=%v", todo.ID, err)
			}
		} else if daysUntilDue <= 3 {
			if err := u.sendNotificationIfNeeded(ctx, todo, entity.NotificationTypeApproaching, now); err != nil {
				log.Printf("[WARN] approaching通知送信に失敗: todoID=%d, error=%v", todo.ID, err)
			}
		}
	}

	return nil
}

func (u *notificationUsecase) sendNotificationIfNeeded(ctx context.Context, todo *entity.Todo, notifType string, now time.Time) error {
	existing, err := u.notificationRepo.FindByTodoIDAndType(ctx, todo.ID, notifType)
	if err != nil {
		return fmt.Errorf("通知重複チェックに失敗: %w", err)
	}
	if existing != nil {
		return nil
	}

	user, err := u.userRepo.FindByID(ctx, todo.UserID)
	if err != nil {
		return fmt.Errorf("ユーザー取得に失敗: userID=%d, %w", todo.UserID, err)
	}

	subject, body := u.buildEmailContent(todo, notifType)
	if err := u.emailSender.Send(ctx, user.Email, subject, body); err != nil {
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	notification := &entity.Notification{
		TodoID: todo.ID,
		UserID: todo.UserID,
		Type:   notifType,
	}
	if err := u.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("通知レコード作成に失敗: %w", err)
	}

	log.Printf("[INFO] 通知送信完了: todoID=%d, type=%s, userEmail=%s", todo.ID, notifType, user.Email)
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
