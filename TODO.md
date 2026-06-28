# TODO - 1週間スケジュール

## AWS構成

```
[ユーザーのブラウザ]
    │
    │  ① HTML/CSS/JS を取得
    ↓
Route 53 → CloudFront → S3（フロントエンド配信）

[ユーザーのブラウザ上でVue 3アプリが動く]
    │
    │  ② APIリクエスト
    ↓
Route 53 → ALB → EC2（Go APIサーバー）
                    │
                    ↓
                  RDS（MySQL）

[定期バッチ（1日1回）]
EventBridge → Lambda → SES（メール送信）
                │
                ↓
              RDS（MySQL）
```

## Terraform 3リソース群

| # | ファイル | AWSリソース |
|---|---------|------------|
| ① | frontend.tf | S3 + CloudFront |
| ② | backend.tf | EC2 + ALB + RDS |
| ③ | lambda.tf | Lambda + EventBridge + SES |

## スケジュール

### Day1: メール通知機能 バックエンド実装
- [ ] Notification エンティティ (domain/entity/notification.go)
- [ ] Notification リポジトリ インターフェース + 実装
- [ ] Notification ユースケース（approaching / overdue 判定、重複防止）
- [ ] ユニットテスト

### Day2: フロントエンド実装（Vue 3 + Vite）
- [ ] frontend/ にプロジェクト初期化
- [ ] Firebase Auth（ログイン・ログアウト・ユーザー登録）
- [ ] TODO一覧表示、作成、編集、削除のUI
- [ ] バックエンドAPIとの接続

### Day3: フロントエンド完成 + バッチ処理コード実装
- [ ] フロントエンドの仕上げ・動作確認
- [ ] backend/cmd/batch/main.go（Lambdaハンドラー）
- [ ] infrastructure/email/ にSESメール送信処理
- [ ] バッチ処理のユニットテスト

### Day4: Docker化 + docker-compose
- [ ] backend/Dockerfile（Go APIサーバー用）
- [ ] backend/Dockerfile.batch（Lambda用）
- [ ] frontend/Dockerfile（ビルド + nginx配信）
- [ ] docker-compose.yml（API + Frontend + MySQL のローカル開発環境）
- [ ] ローカルで全体通しで動作確認

### Day5: AWS構成確定 + Terraform実装
- [ ] AWS構成の最終決定
- [ ] frontend.tf（S3 + CloudFront）
- [ ] backend.tf（EC2 + ALB + RDS）
- [ ] lambda.tf（Lambda + EventBridge + SES）
- [ ] 共通（VPC、サブネット、セキュリティグループ）

### Day6: CI/CD パイプライン構築（GitHub Actions）
- [ ] backend: テスト → Dockerビルド → EC2デプロイ
- [ ] frontend: ビルド → S3アップロード → CloudFrontキャッシュ無効化
- [ ] lambda: ビルド → Lambdaデプロイ
- [ ] mainブランチへのpushで自動デプロイ

### Day7: 結合テスト + バッファ
- [ ] terraform apply でインフラ構築
- [ ] 全体通しでの動作確認（ユーザー登録 → TODO CRUD → メール通知）
- [ ] 遅延タスクの消化
- [ ] バグ修正・微調整
