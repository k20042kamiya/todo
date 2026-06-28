# CI/CD セットアップガイド

GitHub Actions によるCI/CDパイプラインの初期設定手順です。

## ワークフロー概要

| ワークフロー | トリガー | 処理内容 |
|---|---|---|
| `backend.yml` | `backend/` 変更 → main push | Go test → ECR push → ECS Fargate デプロイ |
| `frontend.yml` | `frontend/` 変更 → main push | 型チェック → Vite build → S3 sync → CloudFront無効化 |
| `lambda.yml` | `lambda/` 変更 → main push | Go test → ECR push → Lambda更新 |

---

## 1. AWS: OIDC プロバイダーの作成

長期アクセスキーを使わず、GitHub Actions が一時的なIAM認証情報を取得する仕組みです。

AWSコンソール → **IAM → ID プロバイダー → プロバイダーを追加**

| 項目 | 値 |
|---|---|
| プロバイダーのタイプ | OpenID Connect |
| プロバイダーのURL | `https://token.actions.githubusercontent.com` |
| 対象者 (Audience) | `sts.amazonaws.com` |

---

## 2. AWS: IAM ロールの作成

### 信頼ポリシー

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_ORG/YOUR_REPO:ref:refs/heads/main"
        }
      }
    }
  ]
}
```

> `YOUR_ACCOUNT_ID` / `YOUR_ORG` / `YOUR_REPO` を実際の値に置き換えてください。

### アタッチするポリシー

| ポリシー | 用途 |
|---|---|
| `AmazonEC2ContainerRegistryPowerUser` | ECRへのpush (backend / lambda) |
| ECS個別ポリシー（下記）| ECS タスク定義の更新・サービスのデプロイ (backend) |
| `CloudFrontFullAccess` | CloudFrontキャッシュ無効化 (frontend) |
| `AWSLambda_FullAccess` | Lambda関数コードの更新 (lambda) |
| S3バケット個別ポリシー（下記）| frontend配信バケットへのsync (frontend) |

ECS個別ポリシー例:
```json
{
  "Effect": "Allow",
  "Action": [
    "ecs:RegisterTaskDefinition",
    "ecs:UpdateService",
    "ecs:DescribeTaskDefinition",
    "ecs:DescribeServices",
    "iam:PassRole"
  ],
  "Resource": "*"
}
```

S3バケット個別ポリシー例:
```json
{
  "Effect": "Allow",
  "Action": ["s3:PutObject", "s3:GetObject", "s3:DeleteObject", "s3:ListBucket"],
  "Resource": [
    "arn:aws:s3:::YOUR_BUCKET_NAME",
    "arn:aws:s3:::YOUR_BUCKET_NAME/*"
  ]
}
```

---

## 3. AWS: ECS タスク実行ロールの確認

Terraform で作成される ECS タスク実行ロール（`ecs_execution`）に以下のポリシーがアタッチされていることを確認する。
環境変数やシークレットは ECS タスク定義に直接定義されているため、EC2 への追加作業は不要。

| ポリシー | 用途 |
|---|---|
| `AmazonECSTaskExecutionRolePolicy` | ECRからのdocker pull・CloudWatch Logsへの書き込み |
| SSM Parameter Store 読み取りポリシー（下記）| DB パスワード等のシークレット取得 |

SSM 読み取りポリシー例:
```json
{
  "Effect": "Allow",
  "Action": ["ssm:GetParameters", "ssm:GetParameter"],
  "Resource": "arn:aws:ssm:ap-northeast-1:YOUR_ACCOUNT_ID:parameter/todo/*"
}
```

---

## 4. GitHub: Secrets / Variables の設定

リポジトリの **Settings → Secrets and variables → Actions** で設定します。

### Secrets（機密情報）

| キー | 値の例 | 説明 |
|---|---|---|
| `AWS_ROLE_ARN` | `arn:aws:iam::123456789:role/github-actions-role` | 手順2で作成したIAMロールARN |

### Variables（非機密な設定値）

| キー | 値の例 | 使用ワークフロー |
|---|---|---|
| `AWS_REGION` | `ap-northeast-1` | 全て |
| `ECR_REPOSITORY_BACKEND` | `todo-backend` | backend |
| `ECS_CLUSTER_NAME` | `todo-cluster` | backend |
| `ECS_SERVICE_NAME` | `todo-api` | backend |
| `ECS_TASK_DEFINITION` | `todo-api` | backend |
| `ECS_CONTAINER_NAME` | `api` | backend |
| `ECR_REPOSITORY_LAMBDA` | `todo-lambda` | lambda |
| `S3_BUCKET_FRONTEND` | `todo-frontend-bucket` | frontend |
| `CLOUDFRONT_DISTRIBUTION_ID` | `E1ABCDEFGHIJKL` | frontend |
| `LAMBDA_FUNCTION_NAME` | `todo-notification-batch` | lambda |

---

## 5. GitHub: Environment の設定 (任意)

デプロイ前に手動承認を入れたい場合は **Settings → Environments → production** を作成し、
**Required reviewers** にレビュアーを追加してください。

承認なしで自動デプロイする場合はそのままで問題ありません。
