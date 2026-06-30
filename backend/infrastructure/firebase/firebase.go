package firebase

import (
	"context"
	"os"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// NewFirebaseAuth は Firebase Auth クライアントを初期化して返す。
// 本番: 環境変数 FIREBASE_SERVICE_ACCOUNT_JSON に JSON の中身を直接渡す（ECS Fargate / Secrets Manager）
// ローカル: GOOGLE_APPLICATION_CREDENTIALS でファイルパスを指定する ADC にフォールバック
func NewFirebaseAuth(ctx context.Context) (*auth.Client, error) {
	var app *fb.App
	var err error

	if json := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); json != "" {
		app, err = fb.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(json)))
	} else {
		app, err = fb.NewApp(ctx, nil)
	}
	if err != nil {
		return nil, err
	}

	return app.Auth(ctx)
}
