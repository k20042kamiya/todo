package firebase

import (
	"context"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

// NewFirebaseAuth は Firebase Auth クライアントを初期化して返す。
// 認証情報は環境変数 GOOGLE_APPLICATION_CREDENTIALS で指定されたサービスアカウントキーから読み込まれる。
func NewFirebaseAuth(ctx context.Context) (*auth.Client, error) {
	app, err := fb.NewApp(ctx, nil)
	if err != nil {
		return nil, err
	}

	return app.Auth(ctx)
}
