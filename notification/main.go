package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"notification/infrastructure/database"
	"notification/infrastructure/email"
	infraRepo "notification/infrastructure/repository"
	"notification/usecase"
)

func run(ctx context.Context) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "unexpected panic", "panic", r)
			retErr = fmt.Errorf("unexpected panic: %v", r)
		}
	}()

	db, err := database.NewDB()
	if err != nil {
		return fmt.Errorf("DB接続に失敗: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("DBインスタンス取得に失敗: %w", err)
	}
	defer sqlDB.Close()

	fromEmail := os.Getenv("SES_FROM_EMAIL")
	emailSender, err := email.NewSesSender(ctx, fromEmail)
	if err != nil {
		return fmt.Errorf("メール送信クライアント初期化に失敗: %w", err)
	}

	notifRepo := infraRepo.NewNotificationRepository(db)
	todoRepo := infraRepo.NewTodoRepository(db)
	userRepo := infraRepo.NewUserRepository(db)
	notifUsecase := usecase.NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender)

	if err := notifUsecase.CheckAndSendNotifications(ctx); err != nil {
		return fmt.Errorf("通知送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "通知バッチ処理が完了しました")
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "バッチ処理が失敗しました", "error", err)
		os.Exit(1)
	}
}
