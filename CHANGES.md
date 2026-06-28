# 変更内容まとめ

デプロ���準備のために行った修正の一覧です。

---

## 1. CI/CD：ECS Fargate に統一

**ファイル：** `.github/workflows/backend.yml`

### 変更前
- EC2 インスタンスに SSM 経由で `docker run` していた
- 必要 Variable：`EC2_INSTANCE_ID`

### 変更後
- ECS サービスのタスク定義を更新してローリングデプロイする方式に変更
- 必要 Variable が以下に変わった

| 旧 | 新 |
|---|---|
| `EC2_INSTANCE_ID` | `ECS_CLUSTER_NAME` |
| — | `ECS_SERVICE_NAME` |
| — | `ECS_TASK_DEFINITION` |
| — | `ECS_CONTAINER_NAME` |

---

## 2. Firebase：ECS Fargate で動く形に修正

**ファイル：** `backend/infrastructure/firebase/firebase.go`

### 変更前
- `GOOGLE_APPLICATION_CREDENTIALS`（ファイルパス）を使う ADC のみ
- Fargate にはファイルを置けないため起動時にクラッシュしていた

### 変更後
- `FIREBASE_SERVICE_ACCOUNT_JSON` 環境変数に JSON の中身が入っていればそれを使う
- 環境変数がなければ ADC にフォールバック（ローカル開発はこれまで通り）

```
本番（ECS）: FIREBASE_SERVICE_ACCOUNT_JSON → Secrets Manager から注入
ローカル:    GOOGLE_APPLICATION_CREDENTIALS → 従来通りファイル指定
```

---

## 3. Firebase：Secrets Manager に格納

**ファイル：** `terraform/backend.tf`、`terraform/iam.tf`、`terraform/variables.tf`

### 追加したリソース
- `aws_secretsmanager_secret`：Firebase サービスアカウント JSON を格納するシークレット
- `aws_secretsmanager_secret_version`：`file()` で JSON ファイルを読んで登録
- `aws_iam_role_policy.ecs_execution_secrets_manager`：ECS 実行ロールに Secrets Manager 読み取り権限を付与

### ECS タスク定義への追加
```
secrets = [
  DB_PASSWORD                   ← SSM Parameter Store（既存）
  FIREBASE_SERVICE_ACCOUNT_JSON ← Secrets Manager（追加）
]
```

---

## 4. ヘルスチェックエンドポイントを追加

**ファイル：** `backend/cmd/main.go`

### 変更前
- `/health` エンドポイントが存在しなかった
- ALB のヘルスチェック設定（`path = "/health"`）と噛み合わず、ECS タスクが常に Unhealthy になっていた

### 変更後
```
GET /health → 200 OK {"status": "ok"}
```

---

## 5. Graceful Shutdown を追加

**ファイル：** `backend/cmd/main.go`

### 変更前
- `e.Logger.Fatal(e.Start(":8080"))` のみ
- ECS がタスク停止時に送る SIGTERM を無視して強制終了 → リクエスト途中で切断されていた

### 変更後
- SIGTERM / SIGINT を受け取ったら 10 秒のタイムアウトで安全にシャットダウン

---

## 6. DB コネクションプールを設定

**ファイル：** `backend/infrastructure/database/database.go`

### 変更前
- コネクションプールの設定なし（Go デフォルト：無制限）

### 変更後
```
MaxIdleConns:    10
MaxOpenConns:   100
ConnMaxLifetime: 1時間
```

---

## 7. エラー判定を型安全に変更

**ファイル：** `backend/todo/usecase.go`、`backend/todo/handler.go`、`backend/todo/repository_impl.go`

### 変更前
- usecase が `errors.New("forbidden")` / `errors.New("record not found")` を返す
- handler が `err.Error() == "forbidden"` のような文字列比較で HTTP ステータスを決めていた
- `shared/errors` パッケージ（`ErrCodeForbidden` 等）が定義されているのに未使用だった

### 変更後
- repository：`gorm.ErrRecordNotFound` → `apperrors.New(ErrCodeNotFound, ...)`
- usecase：`errors.New("forbidden")` → `apperrors.New(ErrCodeForbidden, ...)`
- handler：文字列比較 → `apperrors.GetCode(err).HTTPStatus()` で HTTP ステータスを自動解決

---

## GitHub Actions に追加が必要な Variables

`DEPLOY_CHECKLIST.md` の「3. GitHub Actions Secrets / Variables」も更新済み。
以下の Variables を GitHub リポジトリに追加すること。

| キー | 値の例 |
|---|---|
| `ECS_CLUSTER_NAME` | `todo-cluster` |
| `ECS_SERVICE_NAME` | `todo-api` |
| `ECS_TASK_DEFINITION` | `todo-api` |
| `ECS_CONTAINER_NAME` | `api` |
