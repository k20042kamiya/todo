# Lambda メール通知バッチ 仕様書

## 1. 概要

EventBridgeによって1日1回起動されるAWS Lambdaバッチ処理。
未完了かつ期日が設定されているTODOを対象に、期日の状態に応じてユーザーへメール通知を送信する。

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.21+ |
| 実行基盤 | AWS Lambda |
| 起動トリガー | EventBridge（1日1回） |
| DB | Amazon RDS (MySQL 8.0) |
| メール送信 | Amazon SES |
| ORM | GORM v2 |

---

## 2. 起動フロー

```
EventBridge (cron) → Lambda handler(ctx)
    │
    ├─ DB接続初期化
    ├─ SESクライアント初期化
    ├─ NotificationUsecase.CheckAndSendNotifications(ctx)
    │     │
    │     ├─ 未完了かつ期日あるTODOを全件取得
    │     └─ 各TODOに対して通知判定 → メール送信 → 通知レコード保存
    │
    └─ 完了ログ出力
```

---

## 3. ディレクトリ構成

```
lambda/
├── main.go                                  # Lambdaエントリーポイント
├── go.mod / go.sum
├── domain/
│   ├── entity/
│   │   ├── notification.go                  # Notificationエンティティ・定数
│   │   ├── todo.go                          # Todoエンティティ
│   │   └── user.go                          # Userエンティティ
│   └── repository/
│       ├── notification_repository.go       # NotificationRepositoryインターフェース
│       └── user_repository.go               # UserRepositoryインターフェース
├── usecase/
│   ├── notification_usecase.go              # 通知ユースケース実装
│   └── notification_usecase_test.go         # ユニットテスト
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

## 4. レイヤー構成と責務

### 4.1 エントリーポイント（main.go）

- Lambdaハンドラー関数を定義し `lambda.Start()` に渡す
- DB接続・SESクライアント・各リポジトリ・ユースケースの依存関係を組み立てる（DI）
- 初期化失敗時はエラーを返し、Lambdaの実行を失敗扱いにする
- 処理完了後、SQLDBのClose処理を `defer` で保証する

### 4.2 ドメイン層（domain/）

- ビジネスロジックに関係するエンティティと、リポジトリの抽象インターフェースのみを置く
- 外部依存（DB・AWS SDK等）を一切持たない

### 4.3 ユースケース層（usecase/）

- 通知の判定・送信・重複防止のビジネスロジックを実装する
- DBやSESの具体実装には依存せず、インターフェース経由でのみ外部と通信する

### 4.4 インフラ層（infrastructure/）

- DB接続、GORMリポジトリ実装、SESクライアントを実装する
- ドメイン層のインターフェースを満たす形で実装する

---

## 5. エンティティ定義

### 5.1 Notification

```go
// domain/entity/notification.go

const (
    NotificationTypeApproaching = "approaching" // 期日3日以内
    NotificationTypeOverdue     = "overdue"     // 期日超過
)

type Notification struct {
    ID     int       // PK, AUTO_INCREMENT
    TodoID int       // FK -> todo.id
    UserID int       // FK -> user.id
    Type   string    // "approaching" or "overdue"
    SentAt time.Time // 送信日時（autoCreateTime）
}
```

### 5.2 Todo（参照用）

```go
// domain/entity/todo.go

type Todo struct {
    ID          int
    UserID      int
    Title       string
    Content     *string
    DueDate     *time.Time
    IsCompleted bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt // 論理削除
}
```

### 5.3 User（参照用）

```go
// domain/entity/user.go

type User struct {
    ID          int
    FirebaseUID string     // カラム名: firebase_uid
    Email       string
    Name        string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt // 論理削除
}
```

---

## 6. リポジトリインターフェース仕様

### 6.1 NotificationRepository

```go
type NotificationRepository interface {
    // 対象TODOと通知種別で既存通知を検索する（重複防止用）
    // 存在しない場合は (nil, nil) を返す
    FindByTodoIDAndType(ctx context.Context, todoID int, notifType string) (*entity.Notification, error)

    // 通知レコードを新規作成する
    Create(ctx context.Context, notification *entity.Notification) error

    // 未完了かつ期日が設定されている全TODOを取得する
    // 条件: is_completed = false AND due_date IS NOT NULL AND deleted_at IS NULL
    FindUncompletedTodosWithDueDate(ctx context.Context) ([]*entity.Todo, error)
}
```

### 6.2 UserRepository

```go
type UserRepository interface {
    // IDでユーザーを取得する
    // 存在しない場合はエラーを返す（gorm.ErrRecordNotFound）
    FindByID(ctx context.Context, id int) (*entity.User, error)
}
```

---

## 7. ユースケース仕様

### 7.1 インターフェース

```go
type NotificationUsecase interface {
    CheckAndSendNotifications(ctx context.Context) error
}
```

### 7.2 EmailSenderインターフェース

```go
type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}
```

### 7.3 CheckAndSendNotifications 処理フロー

```
1. NotificationRepository.FindUncompletedTodosWithDueDate() で対象TODO一覧取得
2. 各TODOについて:
   a. DueDateがnilであればスキップ（DBクエリで除外済みだが念のため）
   b. 現在日時との差分から通知種別を判定（後述）
   c. 判定結果が通知対象外ならスキップ
   d. sendNotificationIfNeeded() を呼び出し
      - 既に同種別の通知が存在すればスキップ（重複防止）
      - ユーザー情報を取得してメール送信
      - 送信成功後、通知レコードをDBに保存
3. 各TODOの通知エラーは [WARN] ログに記録し、処理を継続する（全体失敗にしない）
4. 全件処理完了後、nil を返す
```

### 7.4 通知判定ロジック

| 条件 | 通知種別 |
|------|---------|
| `daysUntilDue < 0`（期日が過去） | `overdue` |
| `0 <= daysUntilDue <= 3`（期日まで3日以内） | `approaching` |
| `daysUntilDue > 3`（期日まで4日以上） | 通知なし |

**現在の実装（要修正）:**

```go
daysUntilDue := time.Until(dueDate).Hours() / 24
```

`time.Until()` は現在時刻からの差分を時間で計算するため、
例えば「今日の23:59 → 明日の00:01」のような場合に `daysUntilDue ≈ 0.013` となり
`approaching` として誤判定される可能性がある。

**修正後の正しい実装:**

```go
today := time.Now().Truncate(24 * time.Hour)
dueDay := dueDate.Truncate(24 * time.Hour)
daysUntilDue := int(dueDay.Sub(today).Hours() / 24)
```

日付単位で切り捨て比較することで、時刻の端数による誤判定を防ぐ。

### 7.5 メール本文仕様

#### approaching（期日間近）

| 項目 | 内容 |
|------|------|
| 件名 | `【期日間近】{title}` |
| 本文 | `TODOの期日が近づいています。\n\nタイトル: {title}\n期日: {YYYY-MM-DD}\n\n期日までに完了してください。` |

#### overdue（期日超過）

| 項目 | 内容 |
|------|------|
| 件名 | `【期日超過】{title}` |
| 本文 | `TODOの期日が過ぎています。\n\nタイトル: {title}\n期日: {YYYY-MM-DD}\n\n早急に対応してください。` |

---

## 8. インフラ実装仕様

### 8.1 DB接続（infrastructure/database/database.go）

- 環境変数からDSNを構築し、GORMでMySQL接続を確立する
- `NamingStrategy.SingularTable: true` によりテーブル名は単数形（`todo`, `user`, `notification`）
- `parseTime=True` によりMySQLのDATETIMEをGoの `time.Time` に自動変換する
- タイムゾーンはサーバーのローカル時刻を使用（`loc=Local`）

### 8.2 トランザクション管理（infrastructure/database/transaction.go）

- `database.GetTx(ctx, db)` はcontextにトランザクションが含まれていればそれを返し、なければ通常のDBを返す
- 現在のバッチ処理はトランザクションを使用していない（各TODO単位で独立したDB操作）

### 8.3 SESメール送信（infrastructure/email/ses_sender.go）

- AWS SDK v2 の `ses.Client` を使用
- `config.LoadDefaultConfig(ctx)` でIAMロールから認証情報を自動取得（Lambda実行ロールに `ses:SendEmail` 権限が必要）
- 送信形式: テキストメール（UTF-8）
- HTML本文は未サポート

---

## 9. 環境変数

| 変数名 | 必須 | 説明 | 例 |
|--------|------|------|-----|
| `DB_USER` | YES | MySQLユーザー名 | `todoapp` |
| `DB_PASSWORD` | YES | MySQLパスワード | `secret` |
| `DB_HOST` | YES | MySQLホスト | `todo-db.xxx.ap-northeast-1.rds.amazonaws.com` |
| `DB_PORT` | YES | MySQLポート | `3306` |
| `DB_NAME` | YES | データベース名 | `tododb` |
| `SES_FROM_EMAIL` | YES | 送信元メールアドレス（SES検証済みであること） | `noreply@example.com` |
| `AWS_REGION` | YES | AWSリージョン（SDK自動参照） | `ap-northeast-1` |

---

## 10. エラーハンドリング方針

| レベル | 対象 | 挙動 |
|--------|------|------|
| Fatal（関数全体失敗） | DB接続失敗、SESクライアント初期化失敗 | エラーを返しLambda実行失敗として記録 |
| Fatal（関数全体失敗） | 未完了TODO一覧取得失敗 | エラーを返しLambda実行失敗として記録 |
| Warning（処理継続） | 個別TODOの通知失敗（ユーザー取得失敗、メール送信失敗、通知レコード作成失敗） | `[WARN]` ログを出力し次のTODOの処理を継続 |

**通知送信失敗時に記録するログ情報:**
- `todoID`
- 通知種別（`approaching` / `overdue`）
- エラー内容
- （未実装・要修正）送信先メールアドレス

---

## 11. 既知の不具合・修正事項

### 11.1 daysUntilDue の計算が時間ベースで誤判定のリスクあり

- **場所**: `usecase/notification_usecase.go:52`
- **問題**: `time.Until(dueDate).Hours() / 24` は時刻の端数を含むため、日付境界付近で誤判定が発生しうる
- **修正方針**: 日付レベルで `Truncate(24 * time.Hour)` してから差分を整数日数で計算する

### 11.2 sendNotificationIfNeeded の未使用引数

- **場所**: `usecase/notification_usecase.go:68`
- **問題**: `now time.Time` 引数が定義されているが関数内で使われていない
- **修正方針**: 引数を削除する（通知判定ロジック修正と合わせて対応する場合は注入するか判断）

### 11.3 SES送信失敗時のログに送信先が含まれていない

- **場所**: `usecase/notification_usecase.go:84`
- **問題**: メール送信失敗時のエラーログに送信先メールアドレスが含まれておらず、調査が困難
- **修正方針**: `fmt.Errorf("メール送信に失敗: to=%s, %w", user.Email, err)` のように送信先を含める

---

## 12. テスト仕様

### 12.1 テスト対象

`usecase/notification_usecase_test.go` にてユニットテストを実施。
DB・SES・AWSへの依存はすべてモックで置き換える。

### 12.2 テストケース一覧

| テスト名 | 内容 |
|----------|------|
| `TestCheckAndSendNotifications_Approaching` | 期日48時間後のTODOに `approaching` 通知が送信され、通知レコードが作成されること |
| `TestCheckAndSendNotifications_Overdue` | 期日24時間前のTODOに `overdue` 通知が送信され、通知レコードが作成されること |
| `TestCheckAndSendNotifications_DuplicateSkip` | 既に通知済みのTODOに対して通知が重複送信されないこと |
| `TestCheckAndSendNotifications_NoDueDate` | 期日なしTODOに対して通知が送信されないこと |
| `TestCheckAndSendNotifications_FarFutureDueDate` | 期日が10日後のTODOに対して通知が送信されないこと |

### 12.3 未テストのケース（追加推奨）

| テスト名 | 内容 |
|----------|------|
| `TestCheckAndSendNotifications_EmailSendFailure` | メール送信失敗時に処理が継続し、通知レコードが作成されないこと |
| `TestCheckAndSendNotifications_UserNotFound` | ユーザー取得失敗時に処理が継続すること |
| `TestCheckAndSendNotifications_DBCreateFailure` | 通知レコード作成失敗時に適切にログが記録されること |
| `TestBuildEmailContent_Approaching` | approaching時のメール件名・本文が正しいこと |
| `TestBuildEmailContent_Overdue` | overdue時のメール件名・本文が正しいこと |

---

## 13. テストデータ（結合テスト用）

```sql
-- テスト用ユーザー
INSERT INTO user (firebase_uid, email, name, created_at, updated_at)
VALUES ('test-firebase-uid-001', 'test@example.com', 'テストユーザー', NOW(), NOW());

-- 期日間近（3日以内）のTODO
INSERT INTO todo (user_id, title, content, due_date, is_completed, created_at, updated_at)
VALUES (1, '期日間近テスト', '期日3日以内のTODO', DATE_ADD(CURDATE(), INTERVAL 2 DAY), false, NOW(), NOW());

-- 期日超過のTODO
INSERT INTO todo (user_id, title, content, due_date, is_completed, created_at, updated_at)
VALUES (1, '期日超過テスト', '期日が過ぎたTODO', DATE_SUB(CURDATE(), INTERVAL 1 DAY), false, NOW(), NOW());

-- 通知済みレコード（重複防止の確認用）
-- 上記の期日超過TODOに対してoverdueが送信済みであるとする（todo_idは実際のIDに置き換える）
INSERT INTO notification (todo_id, user_id, type, sent_at)
VALUES (2, 1, 'overdue', NOW());
```

---

## 14. AWSリソース要件

### Lambda実行ロールに必要なIAMポリシー

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ses:SendEmail"],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": "*"
    }
  ]
}
```

> RDSへのアクセスはVPC内での通信のため、LambdaをVPCに配置しセキュリティグループで制御する。

### EventBridgeスケジュール

```
cron(0 0 * * ? *)   # 毎日 00:00 UTC（日本時間 09:00 JST）
```
