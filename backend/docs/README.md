# バックエンド設計書

## ドキュメント一覧

| ファイル | 概要 |
|---------|------|
| [01_auth.md](01_auth.md) | Firebase 認証ミドルウェア |
| [02_todo_get_list.md](02_todo_get_list.md) | `GET /api/v1/todos` TODO一覧取得 |
| [03_todo_create.md](03_todo_create.md) | `POST /api/v1/todos` TODO作成 |
| [04_todo_update.md](04_todo_update.md) | `PUT /api/v1/todos/:id` TODO更新 |
| [05_todo_delete.md](05_todo_delete.md) | `DELETE /api/v1/todos/:id` TODO削除 |

---

## エラー分類の定義

各設計書で使用するエラー分類の共通定義。

### 正常なエラー（ビジネスロジック上想定内）

ユーザーの操作・認証状態・リソースの有無に起因する、設計上起きることが想定されたエラー。

- クライアントに **4xx** を返す
- サーバーサイドログは **出力しない**（セキュリティ情報を残さない）
- 例: バリデーション違反、認証失敗、リソース不在、権限なし

### 異常なエラー（外部システム障害）

DB・Firebase 等の外部システムとの通信障害に起因するエラー。通常は発生しないが、インフラ障害時に発生しうる。

- クライアントに **5xx** を返す（内部情報は隠蔽する）
- サーバーサイドに **`[ERROR]` レベルでログを出力する**
- 例: DB 接続断、クエリ失敗、トランザクション失敗、Firebase 接続失敗

### 想定外のエラー

コードのバグ・nil アクセス・未定義状態等、本来発生してはいけないエラー。

- Echo の `middleware.Recover()` がパニックをキャッチして **500** を返す
- サーバーサイドに **`[ERROR]` レベルでログを出力する**
- 発生した場合はバグとして対処する
- 例: nil ポインタ参照、パニック、context に必須値がない

---

## 共通エラーコード

`shared/errors/codes.go` で定義。

| エラーコード | HTTPステータス | 用途 |
|-------------|--------------|------|
| `NOT_FOUND` | 404 | リソースが存在しない |
| `VALIDATION_ERROR` | 400 | バリデーションエラー |
| `UNAUTHORIZED` | 401 | 認証エラー |
| `FORBIDDEN` | 403 | アクセス権限なし |
| `DATABASE_ERROR` | 500 | DB 処理エラー |
| `INTERNAL_ERROR` | 500 | 予期せぬサーバーエラー |

5xx エラーのメッセージはクライアントへは `"Internal server error"` に統一し、内部詳細を隠蔽する。
