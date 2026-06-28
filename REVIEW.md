# リポジトリレビュー 総合結果

レビュー日: 2026-06-28

---

## Critical（今すぐ対処すべき）

| # | 場所 | 問題 | 対応状況 |
|---|------|------|----------|
| B-3 | `user/usecase.go:22` | `FindOrCreateByFirebaseUID` にトランザクションなし — 同時リクエストで競合状態が発生し 500 | 未対応 |
| F-1 | `router/index.ts:78` | 認証ガードに `authReady` バリアがなく race condition でログイン済みユーザーが弾かれる | 未対応 |
| T-1 | `.gitignore` | `terraform.tfstate` が除外されておらず、誤って `git add .` するとシークレットがコミットされる | **対応済み** |
| T-2 | `versions.tf:15` | S3 リモートバックエンドがコメントアウトのまま — ローカル state で運用中（S3 バケット作成後に手動で有効化が必要） | 未対応（手動作業が必要） |
| T-3 | `backend.tf:81`, `notification.tf:73` | ECS タスクがパブリックサブネットで `assign_public_ip = true` のまま稼働 | 意図的（対応不要） |
| T-4 | `iam.tf:77` | SES IAM ポリシーのアカウント ID が `*` — 他アカウントの SES リソースも対象になりうる | **対応済み** |

---

## Major — Backend（Go）

| # | 場所 | 問題 | 対応状況 |
|---|------|------|----------|
| B-1 | `todo/repository_impl.go:55` | `Delete` に `user_id` スコープなし — usecase の所有権チェックを飛ばすと IDOR | **対応済み** |
| B-2 | `todo/repository_impl.go:50` | `Save` が全カラム上書き — 部分的なエンティティで呼ぶとゼロ値で上書きされる危険 | 未対応 |
| B-4 | `todo/handler.go:142,164` | `err.Error()` をそのままレスポンスに返す — DB テーブル名等が露出 | **対応済み** |
| B-5 | `infrastructure/database/database.go:14` | DSN のパスワードに `@` が含まれると接続先が誤判定される | **対応済み** |
| B-6 | `todo/handler.go:29` + `usecase.go:81` | `content` が `string` — クライアントから NULL クリアができない API 仕様バグ | 未対応 |
| B-7 | `infrastructure/database/transaction.go:29` | `tx.Rollback()` のエラーを破棄 | **対応済み** |
| B-8 | `auth/middleware.go:39` | Firebase 匿名トークン（`email` claim なし）がそのまま通り DB 制約違反で 500 | 未対応 |
| B-9 | `todo/handler.go:97` | タイトルの長さ上限チェックなし | **対応済み（100文字）** |
| B-10 | `auth/middleware.go`（テストなし） | 認証ミドルウェアのテストが皆無 | 未対応 |

---

## Major — Notification（Go）

| # | 場所 | 問題 | 対応状況 |
|---|------|------|----------|
| N-1 | `usecase/notification_usecase.go:52` | `time.Until().Hours()/24` による日付境界の誤判定（SPEC.md 記載済みの既知バグが未修正） | 未対応 |
| N-3 | `usecase/notification_usecase.go:84` | SES 失敗ログに送信先メールが含まれない（既知バグ未修正） | 未対応 |
| N-4 | `infrastructure/repository/user_repository.go:27` | `gorm.ErrRecordNotFound` が素のまま返り、DB 障害と区別できない | 未対応 |
| N-5 | `usecase/notification_usecase.go:83` | メール送信後に DB 保存失敗 → 重複通知。送信前にレコードを先保存すべき | 未対応 |
| N-6 | `infrastructure/database/database.go:13` | `loc=Local` — ECS コンテナは UTC のため RDS(JST) との TZ ズレで時刻誤判定 | 未対応 |

---

## Major — Frontend（Vue）

| # | 場所 | 問題 | 対応状況 |
|---|------|------|----------|
| F-2 | `composables/useAuth.ts:57` | `onAuthStateChanged` singleton vs. guard timing | 未対応 |
| F-3 | `views/TodoView.vue:151` | `handleSave` がエラー時もモーダルを閉じる — ユーザーに保存失敗が伝わらない | 未対応 |
| F-4 | `composables/useTodos.ts:82` | 全 CRUD のエラーが `console.error` で握りつぶされユーザーにフィードバックなし | 未対応 |
| F-5 | `composables/useTodos.ts:172` | `toggleComplete` エラー時にユーザーフィードバックなし | 未対応 |
| F-6 | `composables/useTodoFilter.ts:22` | `useTodos()` を二重呼び出しでシングルトン共有しており設計が脆い | 未対応 |
| F-7 | `components/TodoFormModal.vue:147` | `new Date("2026-06-28")` が UTC midnight に解釈 → JST では期日当日午前 9 時から "overdue" 表示 | 未対応 |

---

## Major — Terraform / CI/CD

| # | 場所 | 問題 | 対応状況 |
|---|------|------|----------|
| T-5 | `backend.tf:3`, `notification.tf:3` | ECR タグが mutable — `latest` が上書き可能でロールバック不可 | 意図的（CI 再実行の利便性優先） |
| T-6 | `backend.tf:44`, `notification.tf:23` | Terraform タスク定義が `:latest` 固定 | 未対応 |
| T-7 | `database.tf:8` | RDS が Single-AZ（`multi_az = true` なし） | 未対応 |
| T-8 | `cloudtrail.tf:47` | CloudTrail が single-region（IAM/STS イベントが記録されない） | 未対応 |
| T-9 | `cloudtrail.tf:4` | CloudTrail バケットに `force_destroy = true` — `terraform destroy` で監査ログが消える | 意図的（個人開発のため手動クリーンアップ不要を優先） |
| T-10 | `cloudtrail.tf` | CloudTrail バケットの暗号化なし | 意図的（AWS デフォルト SSE-S3 で許容、コンプライアンス要件なし） |
| T-11 | `backend.tf:27`, `notification.tf:49` | CloudWatch ロググループが KMS 暗号化なし | 未対応 |
| T-12 | `notification.yml:14` | GitHub Actions ロールに `AmazonECS_FullAccess` を推奨（過剰権限） | **対応済み**（最小権限カスタムポリシーに変更） |
| T-13 | `frontend.yml:16` | GitHub Actions ロールに S3 フルアクセスを推奨（過剰権限） | **対応済み**（最小権限カスタムポリシーに変更） |

---

## Minor

### Backend
| # | 場所 | 問題 |
|---|------|------|
| B-11 | `infrastructure/database/database.go` | `SetConnMaxIdleTime` 未設定 — 長時間アイドル後に stale connection エラー |
| B-12 | `todo/usecase_test.go:25` | `FindByUserID` モックに nil ガードなし — パニックの危険 |
| B-13 | `todo/handler_test.go:25` | `GetTodosByUserID` モックに nil ガードなし |
| B-14 | `auth/middleware.go`（テストなし） | 認証ミドルウェアの重要パスにテストなし |
| B-15 | `user/usecase.go`（テストなし） | `FindOrCreateByFirebaseUID` のエラーパスにテストなし |

### Notification
| # | 場所 | 問題 |
|---|------|------|
| N-2 | `usecase/notification_usecase.go:68` | `now time.Time` 引数が未使用（既知バグ未修正） |
| N-7 | `domain/repository/notification_repository.go:12` | `FindUncompletedTodosWithDueDate` が `NotificationRepository` に混在（Clean Architecture 違反） |
| N-8 | `usecase/notification_usecase.go:100` | `buildEmailContent` の default case がなく未知の type で空メールが送られる |

### Frontend
| # | 場所 | 問題 |
|---|------|------|
| F-8 | `views/LoginView.vue:74` | 認証操作中のローディング状態なし — ダブルサブミット可能 |
| F-9 | `views/TodoView.vue:175` | `handleLogout` がエラーを握りつぶし、失敗時もログアウト済みに見える |
| F-10 | `components/TodoItem.vue:72` | `isOverdue()` / `formatDate()` を computed にすべき |
| F-11 | `router/index.ts:84` | ログイン済みユーザーが `/login` に再アクセスしてもリダイレクトなし |
| F-12 | `lib/api.ts:68` | 401 時のトークンリフレッシュリトライなし |
| F-13 | `composables/useTodos.ts:204` | `removeCompleted` が `Promise.all` — 部分失敗で UI と DB が不整合になる |

### Terraform / CI/CD
| # | 場所 | 問題 |
|---|------|------|
| T-14 | `backend.tf:78` | ECS オートスケーリングなし |
| T-15 | `database.tf` | RDS Performance Insights / Enhanced Monitoring なし |
| T-16 | `locals.tf:6` | `Environment` / `Owner` タグなし — コスト配分が困難 |
| T-17 | `frontend.tf`, `backend.tf` | WAF（WAFv2）が CloudFront / ALB に未設定 |
| T-18 | `notification.tf:69` | Notification SG の egress が `0.0.0.0/0` — 必要なポートのみに絞るべき |
| T-19 | `ses.tf` | SES サンドボックス解除が未ドキュメント |
| T-20 | `frontend.yml:55` | Firebase API キーが Secret ではなく Variable に保存 |

---

## 良かった点

- DB パスワードは SSM SecureString、Firebase 認証情報は Secrets Manager に正しく格納
- RDS は `storage_encrypted = true`・`deletion_protection = true`・バックアップ 7 日設定済み
- ALB → HTTPS リダイレクト、TLS 1.3 設定済み
- S3 フロントエンドバケットは OAC 経由のみで公開アクセスブロック済み
- 3 ワークフローすべてで OIDC 認証（長期アクセスキーなし）を使用
- ECS SG は ALB SG からのみ許可、RDS SG は ECS SG のみに制限
