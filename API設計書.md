# API設計書

## 概要

TODOアプリのバックエンドAPI設計書。Firebase Authentication を用いた認証済みユーザーが自身のTODOを管理するためのREST API。

- **ベースURL**: `http://localhost:8080`
- **APIプレフィックス**: `/api/v1`
- **データ形式**: JSON
- **文字コード**: UTF-8

---

## 認証

`/api/v1/*` 配下の全エンドポイントは Firebase ID Token による認証が必要。

### リクエストヘッダー

```
Authorization: Bearer {Firebase ID Token}
```

### 認証フロー

1. クライアントがFirebase SDKでサインイン → Firebase ID Token取得
2. 全APIリクエストの `Authorization` ヘッダーにBearerトークンをセット
3. サーバーがFirebase Admin SDKでトークンを検証
4. トークンからFirebase UIDを取り出し、`users` テーブルと紐付け
5. 初回アクセス時はFirebase UIDをキーにユーザーを自動作成

### 認証エラー

| 条件 | HTTPステータス | エラーメッセージ |
|------|---------------|-----------------|
| `Authorization` ヘッダーなし | 401 | `"Authorization header is required"` |
| Bearer形式ではない | 401 | `"Bearer token is required"` |
| トークン検証失敗（無効・期限切れ） | 401 | `"Invalid token"` |
| トークンにemailクレームなし | 401 | `"Email is required"` |

---

## 共通仕様

### レスポンス形式

**成功時**: 各エンドポイントの仕様に準ずる

**エラー時**:
```json
{
  "error": "エラーメッセージ"
}
```

### 日時形式

全ての日時はISO 8601形式（UTC）で返却する。

```
2025-01-31T00:00:00Z
```

### HTTPステータスコード

| コード | 用途 |
|--------|------|
| 200 OK | 取得・更新成功 |
| 201 Created | 作成成功 |
| 204 No Content | 削除成功 |
| 400 Bad Request | リクエストパラメータ不正 |
| 401 Unauthorized | 認証エラー |
| 403 Forbidden | アクセス権限なし |
| 404 Not Found | リソースが存在しない |
| 500 Internal Server Error | サーバー内部エラー |

---

## エンドポイント一覧

| メソッド | パス | 認証 | 概要 |
|----------|------|------|------|
| GET | `/health` | 不要 | ヘルスチェック |
| GET | `/api/v1/todos` | 必須 | TODO一覧取得 |
| POST | `/api/v1/todos` | 必須 | TODO作成 |
| PUT | `/api/v1/todos/:id` | 必須 | TODO更新 |
| DELETE | `/api/v1/todos/:id` | 必須 | TODO削除 |

---

## エンドポイント詳細

### ヘルスチェック

```
GET /health
```

サーバーの稼働確認用エンドポイント。認証不要。

**レスポンス** `200 OK`

```json
{
  "status": "ok"
}
```

---

### TODO一覧取得

```
GET /api/v1/todos
```

認証済みユーザー自身のTODO一覧を返す。作成日時の降順。

**レスポンス** `200 OK`

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

TODOが0件の場合は空配列を返す。

```json
{
  "todos": []
}
```

**エラーレスポンス**

| ステータス | 条件 |
|-----------|------|
| 401 | 認証エラー |
| 500 | サーバー内部エラー |

---

### TODO作成

```
POST /api/v1/todos
```

新規TODOを作成する。

**リクエストボディ**

```json
{
  "title": "買い物に行く",
  "content": "牛乳と卵を買う",
  "due_date": "2025-01-31T00:00:00Z"
}
```

**パラメータ仕様**

| フィールド | 型 | 必須 | 制約 | 説明 |
|------------|-----|------|------|------|
| `title` | string | YES | 最大100文字 | タイトル |
| `content` | string | NO | - | 内容（nullを許容） |
| `due_date` | string (ISO 8601) | NO | - | 期日（nullを許容） |

**レスポンス** `201 Created`

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

**エラーレスポンス**

| ステータス | 条件 | メッセージ |
|-----------|------|-----------|
| 400 | `title` が空 | `"Title is required"` |
| 400 | `title` が101文字以上 | `"Title must be 100 characters or less"` |
| 400 | リクエストボディが不正なJSON | `"Invalid request body"` |
| 401 | 認証エラー | `"User not authenticated"` |
| 500 | サーバー内部エラー | `"Failed to create todo"` |

---

### TODO更新

```
PUT /api/v1/todos/:id
```

指定したTODOを更新する。自分が作成したTODOのみ更新可能。

**パスパラメータ**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `id` | integer | TODO ID |

**リクエストボディ**

```json
{
  "title": "買い物に行く（更新）",
  "content": "牛乳と卵と野菜を買う",
  "due_date": "2025-02-01T00:00:00Z",
  "is_completed": true
}
```

**パラメータ仕様**

| フィールド | 型 | 必須 | 制約 | 説明 |
|------------|-----|------|------|------|
| `title` | string | YES | 最大100文字 | タイトル |
| `content` | string | NO | - | 内容（nullを許容） |
| `due_date` | string (ISO 8601) | NO | - | 期日（nullを許容） |
| `is_completed` | boolean | YES | - | 完了フラグ |

**レスポンス** `200 OK`

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

**エラーレスポンス**

| ステータス | 条件 | メッセージ |
|-----------|------|-----------|
| 400 | `id` が数値以外 | `"Invalid todo ID"` |
| 400 | `title` が空 | `"Title is required"` |
| 400 | `title` が101文字以上 | `"Title must be 100 characters or less"` |
| 400 | リクエストボディが不正なJSON | `"Invalid request body"` |
| 401 | 認証エラー | `"User not authenticated"` |
| 403 | 他ユーザーのTODOを更新しようとした | `"Forbidden"` |
| 404 | 指定したTODOが存在しない | `"Not found"` |
| 500 | サーバー内部エラー | `"Internal server error"` |

---

### TODO削除

```
DELETE /api/v1/todos/:id
```

指定したTODOを削除する（論理削除）。自分が作成したTODOのみ削除可能。

**パスパラメータ**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `id` | integer | TODO ID |

**レスポンス** `204 No Content`

レスポンスボディなし。

**エラーレスポンス**

| ステータス | 条件 | メッセージ |
|-----------|------|-----------|
| 400 | `id` が数値以外 | `"Invalid todo ID"` |
| 401 | 認証エラー | `"User not authenticated"` |
| 403 | 他ユーザーのTODOを削除しようとした | `"Forbidden"` |
| 404 | 指定したTODOが存在しない | `"Not found"` |
| 500 | サーバー内部エラー | `"Internal server error"` |

---

## レスポンス型定義

### TodoResponse

```typescript
type TodoResponse = {
  id: number;           // TODO ID
  user_id: number;      // ユーザーID
  title: string;        // タイトル
  content: string | null; // 内容
  due_date: string | null; // 期日 (ISO 8601)
  is_completed: boolean; // 完了フラグ
  created_at: string;   // 作成日時 (ISO 8601)
  updated_at: string;   // 更新日時 (ISO 8601)
};
```

### TodoListResponse

```typescript
type TodoListResponse = {
  todos: TodoResponse[];
};
```

---

## エラーコード仕様

内部エラーコードとHTTPステータスのマッピング。

| エラーコード | HTTPステータス | 概要 |
|-------------|---------------|------|
| `NOT_FOUND` | 404 | リソースが存在しない |
| `VALIDATION_ERROR` | 400 | バリデーションエラー |
| `UNAUTHORIZED` | 401 | 認証エラー |
| `FORBIDDEN` | 403 | アクセス権限なし |
| `DATABASE_ERROR` | 500 | DB処理エラー |
| `INTERNAL_ERROR` | 500 | 予期せぬサーバーエラー |

5xxエラーのメッセージは `"Internal server error"` に統一し、内部詳細をクライアントに漏洩させない。

---

## セキュリティ仕様

- **認証**: Firebase Admin SDKによるID Token検証
- **認可**: ユーザーは自身のTODOのみ操作可能（更新・削除時にowner check）
- **データ分離**: 全取得系クエリはログインユーザーの `user_id` でフィルタリング
- **論理削除**: TODOは物理削除せず `deleted_at` に削除日時をセット
- **エラーメッセージ**: 5xxエラーはサーバー内部情報を返さない
