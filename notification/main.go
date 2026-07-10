package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"notification/infrastructure/database"
	"notification/infrastructure/email"
	infraRepo "notification/infrastructure/repository"
	"notification/usecase"
)

// batchTimeout はバッチ全体の実行時間上限。
// 正常時は数秒で完了するため、ハング時の被害（Fargate課金・翌日実行との並行）を
// この時間で打ち切る。ctxは全レイヤー（GORM/SES）に伝播済みのため全経路に効く。
const batchTimeout = 10 * time.Minute

func run(ctx context.Context) error {
	slog.InfoContext(ctx, "バッチ処理開始")

	db, err := database.NewDB()
	if err != nil {
		return fmt.Errorf("DB接続に失敗: %w", err)
	}
	slog.InfoContext(ctx, "DB接続成功")

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
	slog.InfoContext(ctx, "SESクライアント初期化成功")

	notifRepo := infraRepo.NewNotificationRepository(db)
	todoRepo := infraRepo.NewTodoRepository(db)
	userRepo := infraRepo.NewUserRepository(db)
	frontendURL := os.Getenv("FRONTEND_URL")
	notifUsecase := usecase.NewNotificationUsecase(notifRepo, todoRepo, userRepo, emailSender, frontendURL)

	if err := notifUsecase.CheckAndSendNotifications(ctx); err != nil {
		return fmt.Errorf("通知送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "通知バッチ処理が完了しました")
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), batchTimeout)
	defer cancel()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "バッチ処理が失敗しました", "error", err)
		os.Exit(1)
	}
}
