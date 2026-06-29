# 認証ミドルウェア設計書

## 概要

`/api/v1/*` 配下の全エンドポイントに適用される Firebase ID Token 検証ミドルウェア。
トークン検証・ユーザー自動作成・コンテキストへの userID セットを担う。

- **実装ファイル**: `auth/middleware.go`
- **適用範囲**: `api.Use(auth.Auth(...))`

---

## 処理フロー

```
リクエスト受信
  └─ 1. Authorization ヘッダー取得
  └─ 2. Bearer トークン抽出
  └─ 3. Firebase Admin SDK でトークン検証
  └─ 4. トークンから email / name クレーム取得
  └─ 5. FindOrCreateByFirebaseUID でユーザー取得 or 作成
  └─ 6. context に userID をセット
  └─ 7. 次のハンドラーへ
```

---

## エラーの洗い出し

| # | 発生箇所 | エラー内容 | 発生条件 |
|---|---------|-----------|---------|
| 1 | ヘッダー取得 | Authorization ヘッダーなし | クライアントがヘッダーをセットしていない |
| 2 | トークン抽出 | Bearer 形式でない | `Authorization: <token>` のように Bearer なし |
| 3 | Firebase検証 | トークン無効 | 改ざん・期限切れ・不正な署名 |
| 4 | Firebase検証 | Firebase接続失敗 | Firebase サービスダウン・ネットワーク障害 |
| 5 | クレーム取得 | email クレームなし | 匿名ログイン等、emailのないFirebaseユーザー |
| 6 | DB検索 | ユーザー検索失敗 | DB接続断・クエリエラー |
| 7 | DB作成 | ユーザー作成失敗 | DB接続断・email UNIQUE制約違反（レースコンディション）|
| 8 | DB作成 | トランザクション失敗 | DB接続断 |

---

## エラー分類と処理規定

### 正常なエラー（ビジネスロジック上想定内）

ユーザーの操作・認証状態に起因する。クライアントに 4xx を返す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 1 | Authorization ヘッダーなし | 401 | `{"error": "Authorization header is required"}` | なし |
| 2 | Bearer 形式でない | 401 | `{"error": "Bearer token is required"}` | なし |
| 3 | トークン無効・期限切れ | 401 | `{"error": "Invalid token"}` | なし |
| 5 | email クレームなし | 401 | `{"error": "Email is required"}` | なし |

### 異常なエラー（外部システム障害）

Firebase・DB との通信障害に起因する。クライアントに 5xx を返し、サーバーサイドでログを残す。

| # | エラー内容 | HTTPステータス | レスポンス | ログ |
|---|-----------|--------------|-----------|------|
| 4 | Firebase 接続失敗 | 401 | `{"error": "Invalid token"}` | — ※1 |
| 6 | ユーザー検索失敗 | 500 | `{"error": "Internal server error"}` | `[ERROR] FindOrCreateByFirebaseUID failed` |
| 7 | ユーザー作成失敗 | 500 | `{"error": "Internal server error"}` | `[ERROR] FindOrCreateByFirebaseUID failed` |

> ※1 Firebase SDK が `VerifyIDToken` のエラーとして返すため、クライアントには Invalid token として応答する。サーバーログは Firebase SDK 内部のみ。必要に応じて `log.Printf` を追加する。

### 想定外のエラー

コードのバグや nil アクセス等、発生してはいけないエラー。Echo の `middleware.Recover()` がパニックをキャッチして 500 を返す。

| エラー内容 | HTTPステータス | レスポンス | ログ |
|-----------|--------------|-----------|------|
| パニック・nil ポインタ等 | 500 | `{"message": "Internal Server Error"}` | Echo が自動出力 |

---

## レスポンス仕様

### 成功時

レスポンスボディなし。次のハンドラーに制御を渡す。

### エラー時

```json
{
  "error": "エラーメッセージ"
}
```

---

## 現状の実装課題

| 課題 | 内容 | 対応方針 |
|------|------|---------|
| Firebase 接続失敗のログなし | `VerifyIDToken` 失敗時にサーバーログが残らない | `log.Printf("[ERROR] VerifyIDToken failed: %v", err)` を追加する |
| email UNIQUE 制約違反 | レースコンディションで同一 email のユーザーが同時作成されると CREATE 失敗する | `FindOrCreate` をトランザクション + SELECT FOR UPDATE にする、またはリトライ処理を追加する |
