# TODOアプリ

Firebase Authentication で認証済みユーザーが TODO を管理するWebアプリ。  
期日が近づいた・超過した TODO をメールで通知するバッチ機能を持つ。

## ディレクトリ構成

```
todo/
├── backend/        # Go / Echo APIサーバー
├── frontend/       # Vue.js フロントエンド
├── notification/   # メール通知バッチ（Go）
├── terraform/      # AWSインフラ（Terraform）
├── docs/           # 設計書
└── docker-compose.yml
```

## ローカル起動

### 前提

- Docker / Docker Compose
- Firebase プロジェクトの Admin SDK 秘密鍵（JSON）

### 手順

1. Firebase Admin SDK の秘密鍵を `backend/` 直下に配置する

2. 起動

```bash
docker compose up --build
```

| サービス | URL |
|---------|-----|
| フロントエンド | http://localhost:3000 |
| バックエンドAPI | http://localhost:8080 |
| MySQL | localhost:3306 |

### 個別起動

**バックエンドのみ**

```bash
cd backend
go run ./cmd/main.go
```

**フロントエンドのみ**

```bash
cd frontend
npm install
npm run dev
```

## 環境変数

### バックエンドAPI

| 変数名 | 説明 |
|--------|------|
| `DB_HOST` | MySQLホスト |
| `DB_PORT` | MySQLポート（3306） |
| `DB_USER` | MySQLユーザー名 |
| `DB_PASSWORD` | MySQLパスワード |
| `DB_NAME` | データベース名 |
| `GOOGLE_APPLICATION_CREDENTIALS` | Firebase Admin SDK 秘密鍵のパス |

### 通知バッチ

| 変数名 | 説明 |
|--------|------|
| `DB_HOST` | MySQLホスト |
| `DB_PORT` | MySQLポート |
| `DB_USER` | MySQLユーザー名 |
| `DB_PASSWORD` | MySQLパスワード |
| `DB_NAME` | データベース名 |
| `SES_FROM_EMAIL` | 送信元メールアドレス（SES検証済み） |
| `AWS_REGION_NAME` | AWSリージョン |
| `FRONTEND_URL` | メール本文に埋め込むフロントエンドURL |

## ドキュメント

| ドキュメント | 内容 |
|-------------|------|
| [設計書](docs/設計書.md) | システム概要・機能要件・DB設計 |
| [API設計書](docs/API設計書.md) | エンドポイント仕様・認証・エラーコード |
| [バックエンド アーキテクチャ設計書](backend/README.md) | Clean Architecture のレイヤー構成 |
| [メール通知バッチ 設計書](notification/README.md) | バッチの起動フロー・エラーハンドリング |
| [定期メール送信アーキテクチャ比較](docs/EMAIL_SCHEDULER_ARCHITECTURE.md) | Lambda vs ECS RunTask の採用検討記録 |
