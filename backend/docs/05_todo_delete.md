# TODO削除 API 設計書

## 概要

認証済みユーザーが自身の TODO を論理削除する。他ユーザーの TODO は削除不可。

- **エンドポイント**: `DELETE /api/v1/todos/:id`
- **認証**: 必須（Firebase Bearer Token）
- **削除方式**: 論理削除（`deleted_at` に削除日時をセット）
- **実装ファイル**: `todo/handler.go`, `todo/usecase.go`, `todo/repository_impl.go`

---

## 処理フロー

```
DELETE /api/v1/todos/:id
  └─ [middleware] Firebase Token 検証・userID セット
  └─ [handler] DeleteTodo
        └─ context から userID 取得
        └─ パスパラメータ :id を int に変換
        └─ [usecase] DeleteTodo(ctx, userID, todoID)
              └─ [txManager] トランザクション開始
                    └─ [repository] FindByID(ctx, todoID)
                          └─ SELECT * FROM todos WHERE id = ? LIMIT 1
                    └─ オーナーチェック（todo.UserID == userID）
                    └─ [repository] Delete(ctx, todoID, userID)
                          └─ UPDATE todos SET deleted_at = NOW() WHERE id = ? AND user_id = ?
              └─ トランザクション コミット
  └─ 204 No Content を返す
```

---

## リクエスト仕様

### パスパラメータ

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `id` | integer | YES | TODO ID |

リクエストボディなし。

---

## エラーの洗い出し

| # | 発生レイヤー | エラー内容 | 発生条件 |
|---|------------|-----------|---------|
| 1 | middleware | 認証エラー各種 | → `01_auth.md` 参照 |
| 2 | handler | context に userID なし | middleware がセットに失敗 |
| 3 | handler | :id が数値でない | パスパラメータが文字列等 |
| 4 | txManager | トランザクション開始失敗 | DB 接続断 |
| 5 | repository | FindByID: レコードなし | 指定 ID の TODO が存在しない・論理削除済み |
| 6 | repository | FindByID: DB エラー | DB 接続断・クエリタイムアウト |
| 7 | usecase | オーナーチェック失敗 | todo.UserID ≠ userID（他ユーザーの TODO） |
| 8 | repository | Delete: DB エラー | DB 接続断・制約違反 |
| 9 | txManager | コミット失敗 | DB 接続断・デッドロック |
| 10 | txManager | ロールバック失敗 | DB 接続断（`#6`, `#7`, `#8` のエラー処理中） |
| 11 | repository | Delete: 0件削除 | FindByID 後に別トランザクションが先に削除した場合 |

---

## エラー分類と処理規定

### 正常なエラー（ビジネスロジック上想定内）

ユーザーの操作・権限・リソース不在に起因する。クライアントに 4xx を返す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 3 | :id が数値でない | 400 | `{"error": "Invalid todo ID"}` | なし |
| 5 | TODO が存在しない | 404 | `{"error": "todo not found"}` | なし |
| 7 | 他ユーザーの TODO | 403 | `{"error": "forbidden"}` | なし |

### 異常なエラー（外部システム障害）

DB トランザクションに起因する。クライアントに 5xx を返し、ログを残す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 4 | トランザクション開始失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: DeleteTodo failed, userID=N, todoID=N` |
| 6 | FindByID DB エラー | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: DeleteTodo failed, userID=N, todoID=N` |
| 8 | Delete DB エラー | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: DeleteTodo failed, userID=N, todoID=N` |
| 9 | コミット失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: DeleteTodo failed, userID=N, todoID=N` |
| 10 | ロールバック失敗 | 500 | `{"error": "Internal server error"}` | `slog.ErrorContext: transaction rollback failed` |

### 想定外のエラー

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 2 | context に userID なし | 401 | `{"error": "User not authenticated"}` | なし |
| 11 | 0件削除（TOCTOU） | 現状エラーにならない（204 を返す） | — | なし |

> `#11` について: `FindByID` で存在確認後、`Delete` で 0 件更新になっても現在の実装はエラーとしない。実害はないが厳密にする場合は `RowsAffected == 0` のチェックを追加する。

---

## レスポンス仕様

### 成功時 `204 No Content`

レスポンスボディなし。

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
| TOCTOU（検索→削除の間隔）| `FindByID` 後に別トランザクションが同 TODO を削除した場合、`Delete` が 0 件更新になるが現状は正常扱い | `RowsAffected == 0` の場合に `ErrCodeNotFound` を返すよう修正する |
