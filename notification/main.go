package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"notification/infrastructure/database"
	"notification/infrastructure/email"
	infraRepo "notification/infrastructure/repository"
	"notification/usecase"
)

func run(ctx context.Context) error {
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
	userRepo := infraRepo.NewUserRepository(db)
	notifUsecase := usecase.NewNotificationUsecase(notifRepo, userRepo, emailSender)

	if err := notifUsecase.CheckAndSendNotifications(ctx); err != nil {
		return fmt.Errorf("通知送信に失敗: %w", err)
	}

	log.Println("通知バッチ処理が完了しました")
	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("エラー: %v", err)
	}
}
