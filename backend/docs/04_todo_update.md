# TODO更新 API 設計書

## 概要

認証済みユーザーが自身の TODO を更新する。他ユーザーの TODO は更新不可。

- **エンドポイント**: `PUT /api/v1/todos/:id`
- **認証**: 必須（Firebase Bearer Token）
- **実装ファイル**: `todo/handler.go`, `todo/usecase.go`, `todo/repository_impl.go`

---

## 処理フロー

```
PUT /api/v1/todos/:id
  └─ [middleware] Firebase Token 検証・userID セット
  └─ [handler] UpdateTodo
        └─ context から userID 取得
        └─ パスパラメータ :id を int に変換
        └─ リクエストボディをバインド
        └─ バリデーション（title 必須・最大100文字）
        └─ [usecase] UpdateTodo(ctx, userID, todoID, UpdateInput)
              └─ [txManager] トランザクション開始
                    └─ [repository] FindByID(ctx, todoID)
                          └─ SELECT * FROM todos WHERE id = ? LIMIT 1
                    └─ オーナーチェック（todo.UserID == userID）
                    └─ [repository] Update(ctx, todo)
                          └─ UPDATE todos SET ... WHERE id = ?
              └─ トランザクション コミット
  └─ 更新後の TodoResponse を返す
```

---

## リクエスト仕様

### パスパラメータ

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `id` | integer | YES | TODO ID |

### リクエストボディ

```json
{
  "title": "買い物に行く（更新）",
  "content": "牛乳と卵と野菜を買う",
  "due_date": "2025-02-01T00:00:00Z",
  "is_completed": true
}
```

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `title` | string | YES | 最大100文字 | タイトル |
| `content` | string | NO | — | 内容（null 許容） |
| `due_date` | string (ISO 8601) | NO | — | 期日（null 許容） |
| `is_completed` | boolean | YES | — | 完了フラグ |

---

## エラーの洗い出し

| # | 発生レイヤー | エラー内容 | 発生条件 |
|---|------------|-----------|---------|
| 1 | middleware | 認証エラー各種 | → `01_auth.md` 参照 |
| 2 | handler | context に userID なし | middleware がセットに失敗 |
| 3 | handler | :id が数値でない | パスパラメータが文字列等 |
| 4 | handler | JSON バインド失敗 | 不正な JSON 形式・型不一致 |
| 5 | handler | title が空文字 | `title` フィールドなし or 空文字列 |
| 6 | handler | title が 101 文字以上 | バリデーション違反 |
| 7 | txManager | トランザクション開始失敗 | DB 接続断 |
| 8 | repository | FindByID: レコードなし | 指定 ID の TODO が存在しない・論理削除済み |
| 9 | repository | FindByID: DB エラー | DB 接続断・クエリタイムアウト |
| 10 | usecase | オーナーチェック失敗 | todo.UserID ≠ userID（他ユーザーの TODO） |
| 11 | repository | Update: DB エラー | DB 接続断・制約違反 |
| 12 | txManager | コミット失敗 | DB 接続断・デッドロック |
| 13 | txManager | ロールバック失敗 | DB 接続断（`#9`, `#10`, `#11` のエラー処理中） |

---

## エラー分類と処理規定

### 正常なエラー（ビジネスロジック上想定内）

ユーザーの入力誤り・権限・リソース不在に起因する。クライアントに 4xx を返す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 3 | :id が数値でない | 400 | `{"error": "Invalid todo ID"}` | なし |
| 4 | JSON バインド失敗 | 400 | `{"error": "Invalid request body"}` | なし |
| 5 | title が空文字 | 400 | `{"error": "Title is required"}` | なし |
| 6 | title が 101 文字以上 | 400 | `{"error": "Title must be 100 characters or less"}` | なし |
| 8 | TODO が存在しない | 404 | `{"error": "todo not found"}` | なし |
| 10 | 他ユーザーの TODO | 403 | `{"error": "forbidden"}` | なし |

### 異常なエラー（外部システム障害）

DB トランザクションに起因する。クライアントに 5xx を返し、ログを残す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 7 | トランザクション開始失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: UpdateTodo failed, userID=N, todoID=N` |
| 9 | FindByID DB エラー | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: UpdateTodo failed, userID=N, todoID=N` |
| 11 | Update DB エラー | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: UpdateTodo failed, userID=N, todoID=N` |
| 12 | コミット失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: UpdateTodo failed, userID=N, todoID=N` |
| 13 | ロールバック失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: transaction rollback failed` |

### 想定外のエラー

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 2 | context に userID なし | 401 | `{"error": "User not authenticated"}` | なし |

---

## レスポンス仕様

### 成功時 `200 OK`

```json
{
  "id": 1,
  "user_id": 1,
  "title": "買い物に行く（更新）",
  "content": "牛乳と卵と野菜を買う",
  "due_date": "2025-02-01T00:00:00Z",
  "is_completed": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-02T00:00:00Z"
}
```

### エラー時

```json
{
  "error": "エラーメッセージ"
}
```

---

## 実装課題

| 課題 | 内容 | 対応方針 |
|------|------|---------|
| 楽観ロックなし | 同一 TODO を複数クライアントが同時更新した場合、後勝ちになる | 必要に応じて `updated_at` を条件にした楽観ロックを検討する |
