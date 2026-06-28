# Terraform インフラ レビュー記録

## 概要

AWS の各種スキル（`aws-iam`, `aws-containers`, `securing-s3-buckets`）を使用して Terraform コードをレビューし、
セキュリティ・信頼性・コスト効率の観点から改善を実施した記録。

レビュー実施日: 2026-06-28

---

## インフラ構成の全体像

```
フロントエンド:
  ユーザー → CloudFront → S3（Vue ビルド済み静的ファイル）

バックエンド API:
  ユーザー → ALB (HTTPS 443) → ECS Fargate (Go API, port 8080) → RDS MySQL

通知バッチ:
  EventBridge Scheduler (毎日 09:00 JST) → ECS Fargate (通知バッチ) → RDS → SES

シークレット管理:
  DB パスワード → SSM Parameter Store (SecureString)
  Firebase サービスアカウント JSON → Secrets Manager

DNS / 証明書:
  Route53 → ACM (CloudFront 用: us-east-1 / ALB 用: ap-northeast-1)
```

---

## 変更一覧

### 1. Firebase JSON を `var` 経由に統一

**ファイル:** `terraform/backend.tf:142`

**変更前:**
```hcl
secret_string = file("${path.module}/../backend/todo-8a6ad-firebase-adminsdk-fbsvc-cc3c7e2756.json")
```

**変更後:**
```hcl
secret_string = var.firebase_service_account_json
```

**このリソースは何をしているか:**  
Firebase Admin SDK のサービスアカウント JSON を AWS Secrets Manager に保存するリソース。
ECS タスクが起動時にこのシークレットを取得し、Firebase の認証に使用する。

**変更理由:**  
`variables.tf` に `firebase_service_account_json` 変数が定義済みで、`terraform.tfvars` にも値が設定されているにもかかわらず、
`file()` 関数でローカルファイルを直接読み込む実装が混在していた。

- `terraform.tfvars` は `.gitignore` で除外されているが、`file()` で参照しているファイルは git に含まれるリスクがある
- `.gitignore` に `backend/*adminsdk*.json` が入っているため現状は問題ないが、設定の二重管理は混乱を招く
- `var.firebase_service_account_json` を使うことで値の管理を `terraform.tfvars` 一元化できる

**デメリット:** 特になし

---

### 2. IAM: SES 送信権限のリソースを特定ドメインに絞る

**ファイル:** `terraform/iam.tf:75`

**変更前:**
```hcl
Action   = ["ses:SendEmail", "ses:SendRawEmail"]
Resource = "*"
```

**変更後:**
```hcl
Action = ["ses:SendEmail", "ses:SendRawEmail"]
Resource = [
  "arn:aws:ses:${var.aws_region}:*:identity/${var.domain_name}",
  "arn:aws:ses:${var.aws_region}:*:identity/${var.ses_sender_email}",
]
```

**このリソースは何をしているか:**  
ECS タスクロール（アプリが実行時に使用する IAM ロール）に SES のメール送信権限を付与するポリシー。
通知バッチが SES を通じてメールを送信するために必要。

**変更理由:**  
`Resource = "*"` は AWS アカウント内のすべての SES Identity からの送信を許可してしまう。
最小権限の原則（Principle of Least Privilege）に従い、使用するドメインとメールアドレスのみに絞る。

SES の ARN 形式:
```
arn:aws:ses:{リージョン}:{アカウントID}:identity/{ドメインまたはメールアドレス}
```

- `identity/tod-oapp.com` → ドメイン単位の Identity
- `identity/noreply@tod-oapp.com` → メールアドレス単位の Identity（SES は両方を別リソースとして扱う）

**メリット:** 将来別ドメインの SES Identity を追加した場合でも、このロールからは送信不可  
**デメリット:** 送信元を増やす場合は IAM ポリシーも更新が必要

---

### 3. ECS サービスに Circuit Breaker を追加

**ファイル:** `terraform/backend.tf`

**追加内容:**
```hcl
deployment_circuit_breaker {
  enable   = true
  rollback = true
}
```

**このリソースは何をしているか:**  
ECS サービスのデプロイ設定。新しいタスクの起動に繰り返し失敗した場合に自動的に前のバージョンへロールバックする仕組み。

**変更理由:**  
Circuit Breaker なしの場合、デプロイ失敗時にタスクが起動と終了を 30 分以上繰り返し続ける。
その間サービス自体は古いタスクで動き続けるが、「デプロイが失敗していること」の検知が遅れる。

| | Circuit Breaker なし | Circuit Breaker あり |
|---|---|---|
| 失敗検知 | 30分以上かかる | 数分で検知 |
| 復旧 | 手動で前バージョンに戻す | 自動ロールバック |
| サービス影響 | なし | なし |

**メリット:** デプロイ失敗の即時検知・自動復旧  
**デメリット:** 特になし

---

### 4. ECS サービスに `health_check_grace_period_seconds` を追加

**ファイル:** `terraform/backend.tf`

**追加内容:**
```hcl
health_check_grace_period_seconds = 60
```

**このリソースは何をしているか:**  
ECS タスクが起動してから ALB がヘルスチェックを開始するまでの猶予時間（秒）。

**変更理由:**  
Go アプリの起動には DB 接続・初期化処理などで数秒〜数十秒かかる。
猶予時間なしでは ALB がアプリ起動前にヘルスチェックを実行し、unhealthy と判定してタスクを強制終了してしまう。
Circuit Breaker（#3）を追加したことで、この誤判定がロールバックのトリガーになるリスクも増えるため、セットで必要な設定。

```
猶予なし:
  タスク起動 → ALB が即ヘルスチェック → 応答なし → unhealthy → タスク強制終了
              → Circuit Breaker 発動 → ロールバック（実際はバグがないのに）

60秒の猶予あり:
  タスク起動 → 60秒待機 → ALB がヘルスチェック → 起動完了済み → healthy → デプロイ成功
```

**メリット:** 起動の遅いアプリでも誤ロールバックを防げる  
**デメリット:** デプロイ完了が 60 秒分遅くなる（許容範囲）

---

### 5. Secrets Manager の回復ウィンドウを 0 → 7 日に変更

**ファイル:** `terraform/backend.tf:142`

**変更前:**
```hcl
recovery_window_in_days = 0
```

**変更後:**
```hcl
recovery_window_in_days = 7
```

**このリソースは何をしているか:**  
Firebase サービスアカウント JSON を保存する Secrets Manager のシークレット設定。
`recovery_window_in_days` はシークレット削除時に「完全削除されるまでの猶予期間」を指定する。

**変更理由:**  
`0` は即時削除（回復不能）を意味する。`terraform destroy` を誤実行した場合にシークレットを復元できない。
7 日の猶予を設けることで誤削除から回復可能になる。

- **追加コスト:** 削除後 7 日間のみ $0.40/月 × (7/30) ≈ $0.09 — 実質ゼロ
- **意図的に即時削除したい場合:** AWS CLI で `--force-delete-without-recovery` オプションを使えばいつでも即時削除可能

```bash
aws secretsmanager delete-secret \
  --secret-id todo-app/firebase/service-account \
  --force-delete-without-recovery
```

**メリット:** 誤削除からの回復が可能  
**デメリット:** 実質なし

---

### 6. ECR ライフサイクルポリシーを追加

**ファイル:** `terraform/backend.tf`, `terraform/locals.tf`

**追加内容:**

`locals.tf` にポリシー定義を一元化:
```hcl
ecr_lifecycle_policy = jsonencode({
  rules = [
    {
      rulePriority = 1
      description  = "Delete untagged images after 1 day"
      selection = {
        tagStatus   = "untagged"
        countType   = "sinceImagePushed"
        countUnit   = "days"
        countNumber = 1
      }
      action = { type = "expire" }
    },
    {
      rulePriority = 2
      description  = "Keep last 10 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 10
      }
      action = { type = "expire" }
    }
  ]
})
```

`backend.tf` に `for_each` でリポジトリに適用:
```hcl
resource "aws_ecr_lifecycle_policy" "this" {
  for_each = {
    api          = aws_ecr_repository.api.name
    notification = aws_ecr_repository.notification.name
  }

  repository = each.value
  policy     = local.ecr_lifecycle_policy
}
```

**このリソースは何をしているか:**  
ECR（Docker イメージのレジストリ）に蓄積するイメージを自動削除するルール。

**変更理由:**  
ライフサイクルポリシーがないと、デプロイのたびにイメージが積み重なり ECR のストレージコスト（$0.10/GB/月）が増大し続ける。

ルール設計の判断:
- **ルール 1 (優先度高):** 未タグイメージ（ビルド失敗等のゴミ）を 1 日後に削除
- **ルール 2 (優先度低):** タグあり・なし問わず合計 10 件を超えたら古いものを削除

`tagStatus: "tagged"` を使う場合は `tagPrefixList` に空文字が使えないため、`tagStatus: "any"` を採用。

**for_each を使った理由:**  
api と notification の 2 リポジトリに同じポリシーを適用するため、コードの重複を避けるために `for_each` で一元管理。
ECR リポジトリが増えた場合も map に 1 行追加するだけで対応可能。

**image_tag_mutability について:**  
`MUTABLE`（上書き可能）のまま維持。理由:
- CI/CD は `latest` と `$github.sha` の両方をプッシュしている
- `IMMUTABLE` に変更すると `latest` タグの上書きができなくなり CI/CD が壊れる
- `IMMUTABLE` にする場合は CI/CD から `latest` プッシュを削除する必要があり、今回はコストメリットと天秤にかけて現状維持とした

**メリット:** イメージ蓄積によるストレージコスト増大を防止  
**デメリット:** 10 件を超えた古いイメージは自動削除されるためロールバックの幅が制限される

---

### 7. ALB に HTTP → HTTPS リダイレクトを追加

**ファイル:** `terraform/backend.tf`, `terraform/security_groups.tf`

**追加内容 (security_groups.tf):**
```hcl
ingress {
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]
}
```

**追加内容 (backend.tf):**
```hcl
resource "aws_lb_listener" "http_redirect" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
```

**このリソースは何をしているか:**  
ALB にポート 80（HTTP）のリスナーを追加し、来たリクエストを 443（HTTPS）にリダイレクトする設定。
あわせて ALB のセキュリティグループにポート 80 のインバウンドを許可するルールを追加。

セキュリティグループとリスナーは**セットで機能する:**
- SG のみ: ポート 80 が ALB に届くが、処理するリスナーがない
- リスナーのみ: ALB が処理しようとしても SG にブロックされて届かない

**変更理由:**  
ポート 80 のリスナーがないと `http://api.tod-oapp.com` へのアクセスが接続エラーになる。
追加コストは 0（ALB リスナーは無料）、デメリットもないため追加。

フロントエンド（CloudFront）側はすでに `viewer_protocol_policy = "redirect-to-https"` が設定済みのため対象外。

**メリット:** HTTP でアクセスしたユーザーを自動的に HTTPS に誘導できる  
**デメリット:** 特になし

---

### 8. RDS ストレージタイプを gp2 → gp3 に変更

**ファイル:** `terraform/database.tf`

**変更前:**
```hcl
storage_type = "gp2"
```

**変更後:**
```hcl
storage_type = "gp3"
```

**このリソースは何をしているか:**  
RDS MySQL インスタンスのディスク種別設定。IOPS（1秒間の読み書き回数）に影響する。

**変更理由:**

| | gp2 | gp3 |
|---|---|---|
| IOPS (20GB の場合) | 60 IOPS | 3,000 IOPS（固定・無料） |
| コスト | $0.138/GB/月 | $0.138/GB/月（同じ） |
| スループット | ベースライン 128 MB/s | ベースライン 125 MB/s |

gp3 はコストが変わらずに IOPS が大幅に向上するため、変更しない理由がない。
Todo アプリ規模では体感差はほぼないが、将来的なクエリ増加にも余裕を持てる。

**multi_az について:**  
`multi_az = true` にすると別の AZ にスタンバイを自動作成し、障害時に自動フェイルオーバーする。
ただし料金が 2 倍（db.t3.micro: 月 $21 → $42）になるため、個人プロジェクトではコストを優先して未設定のまま維持。

**メリット:** コスト変わらず IOPS が 50 倍に向上  
**デメリット:** 特になし

---

### 9. ALB ターゲットグループの Deregistration Delay を 90 秒に設定

**ファイル:** `terraform/backend.tf`

**追加内容:**
```hcl
deregistration_delay = 90
```

**このリソースは何をしているか:**  
デプロイ時に古いタスクを ALB から切り離す際の待機時間。
ALB はこの時間だけ古いタスクへの新規リクエスト送信を止め、処理中のリクエストが完了するのを待つ。

**変更理由:**  
デフォルトは 300 秒（5 分）で、デプロイのたびに最低 5 分かかる原因になる。
90 秒に短縮することでデプロイを高速化。

90 秒を選んだ理由:
- 30 秒は短すぎる（長い処理が切断されるリスク）
- 300 秒はデプロイが遅すぎる
- 90 秒は「処理中リクエストの完了を十分待てる」と「デプロイ速度」のバランス

**メリット:** デプロイ時間の短縮  
**デメリット:** 90 秒以上かかる処理は切断される可能性がある（通常の API レスポンスでは問題なし）

---

### 10. CloudTrail を追加

**ファイル:** `terraform/cloudtrail.tf`（新規作成）

**追加内容:**
```hcl
resource "aws_s3_bucket" "cloudtrail" { ... }
resource "aws_s3_bucket_public_access_block" "cloudtrail" { ... }
resource "aws_s3_bucket_policy" "cloudtrail" { ... }  # CloudTrail の書き込み許可
resource "aws_cloudtrail" "main" {
  name                          = "todo-app-trail"
  s3_bucket_name                = aws_s3_bucket.cloudtrail.id
  include_global_service_events = true
  is_multi_region_trail         = false
  enable_log_file_validation    = true
}
```

**このリソースは何をしているか:**  
AWS に対するすべての API 呼び出しを S3 バケットに記録するサービス。

記録される内容の例:
- ECS タスクの起動・停止
- IAM ロールの変更
- RDS の操作
- S3 バケットポリシーの変更
- `terraform apply` の実行（誰がいつ何を変更したか）

記録されない内容:
- アプリ内部のエラー（Go のパニック等）→ CloudWatch Logs で確認
- HTTP 500 エラー → ALB アクセスログで確認
- DB クエリエラー → CloudWatch Logs で確認

**変更理由:**  
管理イベントの記録は **最初の 1 Trail が無料**。S3 ストレージ代のみかかる（月数十円程度）。
セキュリティインシデントや予期しない設定変更が起きた際の原因調査に不可欠。

**メリット:** 「誰がいつ何をしたか」の監査証跡が残る。コストほぼゼロ  
**デメリット:** S3 バケット名はグローバルで一意である必要があるため、名前衝突が起きた場合は apply がエラーになる

---

## スキップした項目と判断理由

### ECS タスクのパブリックサブネット配置

**指摘内容:** ECS タスクがパブリックサブネット + Public IP で起動しており、コンテナがインターネットに露出している。

**判断:** 現状維持（コスト優先）

**理由:**  
プライベートサブネットに移動する場合、インターネットへの通信のために以下が必要になる:
- NAT Gateway: 月 $32〜
- VPC Interface Endpoints（ecr.api, ecr.dkr, logs, ssm, secretsmanager）: 各 $7〜/月

現状の ECS セキュリティグループはインバウンドを ALB SG からの 8080 ポートのみに絞っているため、
パブリック IP があっても実質的にインターネットからは到達不可。SG が防壁として機能している。

個人・小規模プロジェクトではこのトレードオフは許容範囲。

---

### WAF（Web Application Firewall）

**指摘内容:** CloudFront・ALB に WAF が未設定。

**判断:** スキップ

**理由:** WAF WebACL が $5/月 + ルール $1/月 で、ALB と CloudFront 両方に付けると $10〜/月 の追加コスト。
個人の Todo アプリへの攻撃リスクは低く、コストに見合わないと判断。

---

### S3 の DenyInsecureTransport ポリシー

**指摘内容:** フロントエンド S3 バケットに HTTP を拒否するポリシーが未設定。

**判断:** スキップ（既存の防御で十分と判断）

**理由:**  
以下の多層防御が既に機能している:
1. S3 パブリックアクセスブロック → 直接アクセス不可
2. バケットポリシー → CloudFront の特定 ARN（SourceArn 条件付き）のみ許可
3. CloudFront の `viewer_protocol_policy = "redirect-to-https"` → HTTP → HTTPS 強制

`DenyInsecureTransport` は「IAM 権限を持つ人が誤って HTTP で直接 S3 にアクセスする」シナリオへの対策だが、
このバケットは CloudFront 経由以外すべて拒否されているため実害がない。

---

### RDS の multi_az

**指摘内容:** `multi_az = true` にすることで別 AZ にスタンバイを自動作成し HA を実現できる。

**判断:** スキップ

**理由:** `db.t3.micro` で月 $21 → $42 に倍増。個人プロジェクトでは Todo が数分使えなくなる程度のリスクは許容範囲。

---

### VPC Flow Logs

**指摘内容:** VPC 内のネットワーク通信ログが未設定。

**判断:** スキップ

**理由:** ユーザー判断でスキップ。CloudTrail で AWS API 操作の監査ログは確保できている。

---

## コスト概算

### 月額見込み（常時稼働）

| サービス | 月額 |
|---|---|
| RDS MySQL db.t3.micro (20GB gp3) | ~$21 |
| ALB | ~$20 |
| ECS Fargate API (0.25vCPU/0.5GB, 24時間) | ~$9 |
| CloudFront | ~$1-5 |
| Route53 ホストゾーン | ~$0.50 |
| Secrets Manager | ~$0.45 |
| ECS 通知バッチ（1日1回、数分） | ~$0.10 |
| S3 / ECR / CloudTrail / SES | ~$1 |
| **合計** | **~$52-57/月** |

### コスト削減オプション

| 状態 | 月額 |
|---|---|
| 常時稼働 | ~$55 |
| ECS + RDS 停止 | ~$27（ALB は課金継続） |
| ALB も削除 | ~$7 |

**RDS コストの発生タイミング:**
- 起動中: インスタンス代 + ストレージ代
- 停止中: ストレージ代のみ（$2.76/月）
- 削除後: スナップショットのみ（$1.9/月）または $0

---

## terraform.tfvars のセキュリティについて

`terraform.tfvars` には DB パスワードや Firebase サービスアカウントの秘密鍵が含まれるが、
`.gitignore` に `*.tfvars` が設定済みであり、git 追跡されていないことを確認済み。

`terraform.tfvars` の取り扱い:
- git には含めない（.gitignore で除外済み）
- CI/CD では `TF_VAR_` 環境変数か GitHub Secrets 経由で渡す
- ローカル開発時のみファイルとして使用
