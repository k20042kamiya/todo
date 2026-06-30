# TODO一覧取得 API 設計書

## 概要

認証済みユーザー自身の TODO を全件取得する。

- **エンドポイント**: `GET /api/v1/todos`
- **認証**: 必須（Firebase Bearer Token）
- **実装ファイル**: `todo/handler.go`, `todo/usecase.go`, `todo/repository_impl.go`

---

## 処理フロー

```
GET /api/v1/todos
  └─ [middleware] Firebase Token 検証・userID セット
  └─ [handler] GetTodos
        └─ context から userID 取得
        └─ [usecase] GetTodosByUserID(ctx, userID)
              └─ [repository] FindByUserID(ctx, userID)
                    └─ SELECT * FROM todos WHERE user_id = ? ORDER BY created_at DESC
  └─ TodoListResponse を返す
```

---

## エラーの洗い出し

| # | 発生レイヤー | エラー内容 | 発生条件 |
|---|------------|-----------|---------|
| 1 | middleware | 認証エラー各種 | → `01_auth.md` 参照 |
| 2 | handler | context に userID なし | middleware がセットに失敗（通常は発生しない） |
| 3 | repository | DB クエリ失敗 | DB 接続断・クエリタイムアウト |
| 4 | repository | DB 接続取得失敗 | トランザクションコンテキスト破損（通常は発生しない） |

---

## エラー分類と処理規定

### 正常なエラー（ビジネスロジック上想定内）

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| — | 該当なし（TODO が 0 件は正常: `todos: []` を返す） | — | — | — |

> TODO が存在しない場合はエラーではなく空配列レスポンス。

### 異常なエラー（外部システム障害）

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 3 | DB クエリ失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: GetTodosByUserID failed` |

### 想定外のエラー

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 2 | context に userID なし | 401 | `{"error": "User not authenticated"}` | なし |
| 4 | DB接続取得失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: GetTodosByUserID failed` |

> `#2` は middleware が正常に動作していれば発生しない。

---

## レスポンス仕様

### 成功時 `200 OK`

```json
{
  "todos": [
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
  ]
}
```

TODO が 0 件の場合:

```json
{
  "todos": []
}
```

### エラー時

```json
{
  "error": "エラーメッセージ"
}
```

---

