# デプロイチェックリスト

## 1. ~~インフラとCI/CDの不一致~~ ✅ 解決済み

`.github/workflows/backend.yml` を ECS Fargate デプロイに修正済み。

---

## 2. 設定値の未記入

| 項目 | ファイル | 現状 |
|---|---|---|
| `firebase_project_id` | `terraform/terraform.tfvars` | `"your-firebase-project-id"` のまま |
| Firebase APIキー等 | GitHub Actions Variables | 未設定 |

- [ ] Firebase プロジェクトを作成し、`firebase_project_id` を記入する
- [ ] GitHub Actions Variables に Firebase の値を設定する（後述）

---

## 3. GitHub Actions Secrets / Variables の設定

リポジトリの **Settings → Secrets and variables → Actions** で設定する。

### Secrets（機密情報）

| キー | 説明 |
|---|---|
| `AWS_ROLE_ARN` | OIDC用IAMロールARN（例: `arn:aws:iam::123456789:role/github-actions-role`）|

### Variables（非機密な設定値）

| キー | 説明 | 使用ワークフロー |
|---|---|---|
| `AWS_REGION` | AWSリージョン（例: `ap-northeast-1`）| 全て |
| `ECR_REPOSITORY_BACKEND` | ECRリポジトリ名（例: `todo-backend`）| backend |
| `ECS_CLUSTER_NAME` | ECSクラスター名（例: `todo-cluster`）| backend |
| `ECS_SERVICE_NAME` | ECSサービス名（例: `todo-api`）| backend |
| `ECS_TASK_DEFINITION` | タスク定義のfamily名（例: `todo-api`）| backend |
| `ECS_CONTAINER_NAME` | コンテナ名（例: `api`）| backend |
| `S3_BUCKET_FRONTEND` | フロントエンド配信用S3バケット名 | frontend |
| `CLOUDFRONT_DISTRIBUTION_ID` | CloudFrontディストリビューションID | frontend |
| `LAMBDA_FUNCTION_NAME` | Lambda関数名（例: `todo-notification-batch`）| lambda |
| `VITE_FIREBASE_API_KEY` | Firebase APIキー | frontend |
| `VITE_FIREBASE_AUTH_DOMAIN` | Firebase Auth Domain（例: `todo-xxxx.firebaseapp.com`）| frontend |
| `VITE_FIREBASE_PROJECT_ID` | Firebase プロジェクトID | frontend |

- [ ] 全Secrets / Variablesを設定する

---

## 4. AWS側の事前設定

### IAM OIDC プロバイダー

GitHub Actions が一時的なIAM認証情報を取得するために必要。

AWSコンソール → **IAM → ID プロバイダー → プロバイダーを追加**

| 項目 | 値 |
|---|---|
| プロバイダーのタイプ | OpenID Connect |
| プロバイダーURL | `https://token.actions.githubusercontent.com` |
| 対象者 (Audience) | `sts.amazonaws.com` |

- [ ] OIDC プロバイダーを作成する

### IAM ロール

詳細手順は `CICD_SETUP.md` を参照。

- [ ] IAMロールを作成し、ARNを `AWS_ROLE_ARN` Secret に設定する

### SES

- [ ] `noreply@todo-app.dev` の送信元メールアドレスを SES で検証する
- [ ] SES サンドボックス解除を申請する（本番でのメール送信に必要）

---

## 5. ドメイン

Terraform は `todo-app.dev` で Route53 ホストゾーンを作成する。

- [ ] `todo-app.dev` ドメインを取得済みか確認する
- [ ] `terraform apply` 後、Route53 のネームサーバーをドメインレジストラに設定する

---

## 6. セキュリティ対応 ✅ 対応済み

- `.gitignore` に `backend/*adminsdk*.json` が記載済みで Git 未追跡
- `terraform apply` 時に `file()` で読んで AWS Secrets Manager にアップロード
- ECS タスクは起動時に Secrets Manager から取得（サーバーにファイルは不要）

---

## デプロイ手順（上記が揃った後）

```bash
# 1. Terraform でインフラを構築
cd terraform
terraform init
terraform plan
terraform apply

# 2. Route53 のネームサーバーをドメインレジストラに設定
#    （terraform output で確認できる）

# 3. main ブランチに push → GitHub Actions が自動デプロイ
git push origin main
```
