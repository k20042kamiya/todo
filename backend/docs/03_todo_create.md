# TODO作成 API 設計書

## 概要

認証済みユーザーの TODO を新規作成する。

- **エンドポイント**: `POST /api/v1/todos`
- **認証**: 必須（Firebase Bearer Token）
- **実装ファイル**: `todo/handler.go`, `todo/usecase.go`, `todo/repository_impl.go`

---

## 処理フロー

```
POST /api/v1/todos
  └─ [middleware] Firebase Token 検証・userID セット
  └─ [handler] CreateTodo
        └─ context から userID 取得
        └─ リクエストボディをバインド
        └─ バリデーション（title 必須・最大100文字）
        └─ [usecase] CreateTodo(ctx, userID, CreateInput)
              └─ [txManager] トランザクション開始
                    └─ [repository] Create(ctx, todo)
                          └─ INSERT INTO todos ...
              └─ トランザクション コミット
  └─ 作成した TodoResponse を返す
```

---

## リクエスト仕様

### リクエストボディ

```json
{
  "title": "買い物に行く",
  "content": "牛乳と卵を買う",
  "due_date": "2025-01-31T00:00:00Z"
}
```

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `title` | string | YES | 最大100文字 | タイトル |
| `content` | string | NO | — | 内容（null 許容） |
| `due_date` | string (ISO 8601) | NO | — | 期日（null 許容） |

---

## エラーの洗い出し

| # | 発生レイヤー | エラー内容 | 発生条件 |
|---|------------|-----------|---------|
| 1 | middleware | 認証エラー各種 | → `01_auth.md` 参照 |
| 2 | handler | context に userID なし | middleware がセットに失敗 |
| 3 | handler | JSON バインド失敗 | 不正な JSON 形式・型不一致 |
| 4 | handler | title が空文字 | `title` フィールドなし or 空文字列 |
| 5 | handler | title が 101 文字以上 | バリデーション違反 |
| 6 | txManager | トランザクション開始失敗 | DB 接続断 |
| 7 | repository | INSERT 失敗 | DB 接続断・制約違反 |
| 8 | txManager | コミット失敗 | DB 接続断・デッドロック |
| 9 | txManager | ロールバック失敗 | DB 接続断（`#7` のエラー処理中） |

---

## エラー分類と処理規定

### 正常なエラー（ビジネスロジック上想定内）

ユーザーの入力誤りに起因する。クライアントに 4xx を返す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 3 | JSON バインド失敗 | 400 | `{"error": "Invalid request body"}` | なし |
| 4 | title が空文字 | 400 | `{"error": "Title is required"}` | なし |
| 5 | title が 101 文字以上 | 400 | `{"error": "Title must be 100 characters or less"}` | なし |

### 異常なエラー（外部システム障害）

DB トランザクションに起因する。クライアントに 5xx を返し、ログを残す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 6 | トランザクション開始失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: CreateTodo failed, userID=N` |
| 7 | INSERT 失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: CreateTodo failed, userID=N` |
| 8 | コミット失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: CreateTodo failed, userID=N` |
| 9 | ロールバック失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: transaction rollback failed` |

### 想定外のエラー

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 2 | context に userID なし | 401 | `{"error": "User not authenticated"}` | なし |

---

## レスポンス仕様

### 成功時 `201 Created`

```json
{
  "id": 1,
  "user_id": 1,
  "title": "買い物に行く",
  "content": "牛乳と卵を買う",
  "due_date": "2025-01-31T00:00:00Z",
  "is_completed": false,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### エラー時

```json
{
  "error": "エラーメッセージ"
}
```

---

