# メール通知バッチ 設計書

## 概要

未完了かつ期日が設定されているTODOを対象に、期日の状態に応じてユーザーへメール通知を送信するバッチサービス。

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.24+ |
| 実行形態 | ECS Fargate（EventBridge Schedulerで毎日09:00 JST起動） |
| DB | Amazon RDS (MySQL 8.0) |
| メール送信 | Amazon SES |
| ORM | GORM v2 |

---

## 起動フロー

```
EventBridge Scheduler (cron) → ECS Fargate タスク起動
    │
    ├─ DB接続初期化
    ├─ SESクライアント初期化
    ├─ NotificationUsecase.CheckAndSendNotifications(ctx)
    │     │
    │     ├─ 未完了かつ期日あるTODOを全件取得
    │     └─ 各TODOに対して:
    │           ├─ 期日判定（overdue / approaching / 対象外）
    │           ├─ 重複チェック
    │           ├─ ユーザー取得
    │           ├─ メール送信
    │           └─ 通知レコード保存
    │
    └─ 完了ログ出力
```

---

## ディレクトリ構成

```
notification/
├── main.go
├── go.mod / go.sum
├── notification設計書.md
├── domain/
│   ├── entity/
│   │   ├── errors.go                        # ドメインエラー定義（ErrNotFound, ErrInvalidRecipient）
│   │   ├── notification.go                  # Notificationエンティティ・定数
│   │   ├── todo.go                          # Todoエンティティ
│   │   └── user.go                          # Userエンティティ
│   └── repository/
│       ├── notification_repository.go       # NotificationRepositoryインターフェース
│       └── user_repository.go               # UserRepositoryインターフェース
├── usecase/
│   ├── notification_usecase.go              # 通知ユースケース実装
│   └── notification_usecase_test.go
└── infrastructure/
    ├── database/
    │   ├── database.go                      # DB接続初期化
    │   └── transaction.go                   # トランザクションコンテキスト管理
    ├── email/
    │   └── ses_sender.go                    # SESメール送信実装
    └── repository/
        ├── notification_repository.go       # NotificationRepository実装
        └── user_repository.go               # UserRepository実装
```

---

## レイヤー構成と責務

### エントリーポイント（main.go）

- DB接続・SESクライアント・各リポジトリ・ユースケースの依存を組み立て（DI）
- 初期化失敗時はエラーを返し `slog.ErrorContext` + `os.Exit(1)` でプロセス終了
- `defer recover()` でパニックをキャッチし `[ERROR]` ログ後にプロセス終了

### ドメイン層（domain/）

- エンティティ・リポジトリインターフェース・ドメインエラーのみを置く
- 外部依存（DB・AWS SDK等）を一切持たない

### ユースケース層（usecase/）

- 通知判定・送信・重複防止のビジネスロジックを実装
- インターフェース経由でのみ外部と通信する

### インフラ層（infrastructure/）

- DB接続・GORMリポジトリ実装・SESクライアントを実装
- ドメイン層のインターフェースを満たす形で実装する

---

## エンティティ定義

### Notification

```go
const (
    NotificationTypeApproaching = "approaching" // 期日3日以内
    NotificationTypeOverdue     = "overdue"     // 期日超過
)

type Notification struct {
    ID     int
    TodoID int
    UserID int
    Type   string
    SentAt time.Time // autoCreateTime
}
```

### Todo（参照用）

```go
type Todo struct {
    ID          int
    UserID      int
    Title       string
    Content     *string
    DueDate     *time.Time
    IsCompleted bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt
}
```

### User（参照用）

```go
type User struct {
    ID          int
    FirebaseUID string
    Email       string
    Name        string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt
}
```

---

## リポジトリインターフェース仕様

### NotificationRepository

```go
type NotificationRepository interface {
    // 対象TODOと通知種別で既存通知を検索（重複防止用）
    // 存在しない場合は (nil, nil) を返す
    FindByTodoIDAndType(ctx context.Context, todoID int, notifType string) (*entity.Notification, error)

    // 通知レコードを新規作成
    Create(ctx context.Context, notification *entity.Notification) error

    // 未完了かつ期日が設定されている全TODOを取得
    // 条件: is_completed = false AND due_date IS NOT NULL AND deleted_at IS NULL
    FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error)
}
```

### UserRepository

```go
type UserRepository interface {
    // IDでユーザーを取得
    // 存在しない場合は entity.ErrNotFound でラップされたエラーを返す
    FindByID(ctx context.Context, id int) (*entity.User, error)
}
```

---

## ユースケース仕様

### 通知判定ロジック

日付レベルで切り捨て比較し、時刻の端数による誤判定を防ぐ。

```go
today := time.Now().UTC().Truncate(24 * time.Hour)
dueDay := dueDate.UTC().Truncate(24 * time.Hour)
daysUntilDue := int(dueDay.Sub(today).Hours() / 24)
```

| 条件 | 通知種別 |
|------|---------|
| `daysUntilDue < 0`（期日が過去） | `overdue` |
| `0 <= daysUntilDue <= 3`（期日まで3日以内） | `approaching` |
| `daysUntilDue > 3`（期日まで4日以上） | 通知なし（スキップ） |

### sendNotificationIfNeeded 処理順序

**メール送信を先に行い、成功した場合のみレコードを保存する。**
（逆順にすると、メール送信失敗時に記録だけが残り次回実行で永久にスキップされる。）

```
1. FindByTodoIDAndType で重複確認
2. userRepo.FindByID でユーザー取得
3. emailSender.Send でメール送信      ← 先に送信
4. notificationRepo.Create でレコード保存  ← 成功後に保存
```

### メール本文仕様

| 種別 | 件名 | 本文 |
|------|------|------|
| approaching | `【期日間近】{title}` | `TODOの期日が近づいています。\nタイトル: {title}\n期日: {YYYY-MM-DD}\nTODOを確認する: {FRONTEND_URL}\n期日までに完了してください。` |
| overdue | `【期日超過】{title}` | `TODOの期日が過ぎています。\nタイトル: {title}\n期日: {YYYY-MM-DD}\nTODOを確認する: {FRONTEND_URL}\n早急に対応してください。` |

---

## エラーハンドリング設計

### エラーの洗い出し

| # | 発生フェーズ | エラー内容 | 発生条件 |
|---|------------|-----------|---------|
| 1 | 起動 | DB接続失敗 | RDS障害・ネットワーク障害・認証情報誤り |
| 2 | 起動 | SESクライアント初期化失敗 | IAMロール不足・AWS設定誤り |
| 3 | TODO取得 | `FindUncompletedTodosWithDueDate` 失敗 | DB接続断・クエリエラー |
| 4 | 重複確認 | `FindByTodoIDAndType` 失敗 | DB接続断 |
| 5 | ユーザー取得 | `FindByID` 失敗（NOT FOUND） | TODOのUserIDに対応するユーザーが論理削除済み |
| 6 | ユーザー取得 | `FindByID` 失敗（DBエラー） | DB接続断・クエリエラー |
| 7 | メール送信 | `emailSender.Send` 失敗（個別） | 無効なメールアドレス・SESサンドボックス制限 |
| 8 | メール送信 | `emailSender.Send` 失敗（サービス障害） | SESダウン・スロットリング |
| 9 | 通知レコード保存 | `notificationRepo.Create` 失敗 | DB接続断・制約違反 |
| 10 | 全体 | パニック（nilポインタ参照等） | コードのバグ・未初期化の依存 |

### エラー分類と処理規定

#### 正常なエラー（想定内スキップ）

| # | エラー内容 | 処理 | ログ |
|---|-----------|------|------|
| 5 | ユーザーが存在しない（論理削除済み等） | スキップして継続 | `WARN` |
| 7 | SESメール送信失敗（無効アドレス等） | スキップして継続 | `WARN` |

#### 異常なエラー — プロセス終了

処理継続しても全件失敗が確実な障害。早期終了でCloudWatchアラートを確実に発火させる。

| # | エラー内容 | 処理 | ログ |
|---|-----------|------|------|
| 1 | DB接続失敗（起動時） | `os.Exit(1)` でプロセス終了 | `ERROR` |
| 2 | SESクライアント初期化失敗 | `os.Exit(1)` でプロセス終了 | `ERROR` |
| 3 | TODO取得クエリ失敗 | error を返し `main` が `os.Exit(1)` | `ERROR` |
| 8 | SESサービス障害 | error を返し `main` が `os.Exit(1)` | `ERROR` |

#### 異常なエラー — 個別スキップ

| # | エラー内容 | 処理 | ログ | 安全側の理由 |
|---|-----------|------|------|------------|
| 4 | 重複確認クエリ失敗 | スキップして継続 | `WARN` | エラー時に「未送信」と誤判定して重複送信するリスクを避ける |
| 6 | ユーザー取得DBエラー | スキップして継続 | `WARN` | 送信先が確定できないため送信しない |
| 9 | 通知レコード保存失敗 | ログのみ・処理継続 | `ERROR` | メール送信後に発生するため取り消し不可 |

#### 想定外のエラー

| # | エラー内容 | 処理 | ログ |
|---|-----------|------|------|
| 10 | nilポインタ参照・パニック | `defer recover()` でキャッチし `os.Exit(1)` | `ERROR` |

### エラーハンドリング実装チェックリスト

#### 正常なエラー
- [x] ユーザーが存在しない（`entity.ErrNotFound`）時にスキップ + `WARN` ログ出力している
- [x] SES個別アドレスエラー（`entity.ErrInvalidRecipient`）時にスキップ + `WARN` ログ出力している

#### 異常なエラー
- [x] DB接続失敗時に `os.Exit(1)` でプロセス終了している
- [x] SESクライアント初期化失敗時に `os.Exit(1)` でプロセス終了している
- [x] TODO取得失敗時に error を返し `main` が `os.Exit(1)` でプロセス終了している
- [x] 重複確認クエリ失敗時にスキップ（安全側）+ `WARN` ログ出力している
- [x] メール送信後に通知レコードを保存している
- [x] SESサービス障害を個別エラーと区別してプロセス終了している

#### 想定外のエラー
- [x] `main` の `run()` に `defer recover()` がある

#### ログ品質（slog 構造化ログ）
- [x] 個別TODOエラーに `todo_id` フィールドが含まれている
- [x] ユーザー取得エラーに `user_id` フィールドが含まれている
- [x] メール送信エラーに `to` フィールドが含まれている

---

## インフラ実装仕様

### DB接続（infrastructure/database/database.go）

- 環境変数からDSNを構築し、GORMでMySQL接続を確立する
- `NamingStrategy.SingularTable: true` によりテーブル名は単数形（`todo`, `user`, `notification`）
- `parseTime=True` によりMySQLのDATETIMEをGoの `time.Time` に自動変換する

### SESメール送信（infrastructure/email/ses_sender.go）

- AWS SDK v2 の `ses.Client` を使用
- `config.LoadDefaultConfig(ctx)` でIAMロールから認証情報を自動取得
- `*types.MessageRejected` は `entity.ErrInvalidRecipient` でラップして返す（個別エラーとサービス障害を区別）
- 送信形式: テキストメール（UTF-8）

---

## 環境変数

| 変数名 | 必須 | 説明 |
|--------|------|------|
| `DB_USER` | YES | MySQLユーザー名 |
| `DB_PASSWORD` | YES | MySQLパスワード（SSMから注入） |
| `DB_HOST` | YES | MySQLホスト |
| `DB_PORT` | YES | MySQLポート（3306） |
| `DB_NAME` | YES | データベース名 |
| `SES_FROM_EMAIL` | YES | 送信元メールアドレス（SES検証済みであること） |
| `AWS_REGION_NAME` | YES | AWSリージョン |
| `FRONTEND_URL` | YES | メール本文に埋め込むフロントエンドURL（例: `https://tod-oapp.com`） |

---

## AWSリソース要件

### ECSタスクロールに必要なIAMポリシー

```json
{
  "Statement": [
    { "Effect": "Allow", "Action": ["ses:SendEmail"], "Resource": "*" }
  ]
}
```

### EventBridgeスケジュール

```
cron(0 0 * * ? *)   # 毎日 00:00 UTC（日本時間 09:00 JST）
```

---

## テスト仕様

### テスト対象

`usecase/notification_usecase_test.go` にてユニットテストを実施。
DB・SES・AWSへの依存はすべてモックで置き換える。

### 実装済みテストケース

| テスト名 | 確認内容 |
|----------|---------|
| `TestCheckAndSendNotifications_Approaching` | 期日48時間後のTODOに `approaching` 通知が送信・記録されること |
| `TestCheckAndSendNotifications_Overdue` | 期日24時間前のTODOに `overdue` 通知が送信・記録されること |
| `TestCheckAndSendNotifications_DuplicateSkip` | 送信済みTODOに対して通知が重複送信されないこと |
| `TestCheckAndSendNotifications_NoDueDate` | 期日なしTODOに対して通知が送信されないこと |
| `TestCheckAndSendNotifications_FarFutureDueDate` | 期日が10日後のTODOに対して通知が送信されないこと |

### 追加推奨テストケース

| テスト名 | 確認内容 |
|----------|---------|
| `TestCheckAndSendNotifications_UserNotFound` | ユーザーが存在しない場合にスキップして処理継続すること |
| `TestCheckAndSendNotifications_InvalidRecipient` | SES個別アドレスエラー時にスキップして処理継続すること |
| `TestCheckAndSendNotifications_SESServiceError` | SESサービス障害時に全体処理がエラーで終了すること |
| `TestCheckAndSendNotifications_DuplicateCheckDBError` | 重複確認DBエラー時にスキップ（安全側）して処理継続すること |
| `TestCheckAndSendNotifications_RecordSaveFailure` | レコード保存失敗時にログのみ出力して処理継続すること |
