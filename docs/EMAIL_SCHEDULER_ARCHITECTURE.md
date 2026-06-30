# 定期メール送信アーキテクチャ比較

## 前提

### 要件

- EventBridge Scheduler により定期実行
- RDS からメール送信対象データを取得
- メール本文を生成
- SES でメール送信
- NAT Gateway は構築しない方針
- RDS はプライベートサブネットに配置

---

## 案1：Scheduler → Lambda → SES

### 構成

```
EventBridge Scheduler
        │
        ▼
Lambda（Private Subnet）
        │
        ├── RDS
        │
        ▼
SES
```

Lambda は RDS へ接続するため、VPC 内（プライベートサブネット）へ配置する。

そのため SES へ接続するためには以下のいずれかが必要となる。

- NAT Gateway
- SES VPC Interface Endpoint

### メリット

- 構成が非常にシンプル
- Lambda のみで完結する
- 起動時間が短い
- 運用が容易
- Lambda は実行時間課金のため基本的な実行コストは低い

### デメリット

- NAT Gateway または SES VPC Endpoint が必要
- NAT Gateway を利用する場合、固定費が高い
- Lambda の実行時間・メモリ制限を受ける
- 重いバッチ処理には向かない

---

## 案2：Scheduler → ECS RunTask → SES

### 構成

```
EventBridge Scheduler
        │
        ▼
ECS RunTask（Public Subnet）
        │
        ├── RDS（Private）
        │
        ▼
SES（Internet）
```

RunTask は Public Subnet へ配置し、Public IP を付与する。

RDS は同一 VPC 内通信となるため Private Subnet へ接続可能。

SES はインターネット経由で接続するため NAT Gateway は不要。

### メリット

- NAT Gateway 不要
- SES VPC Endpoint 不要
- Go アプリをそのまま利用できる
- Lambda の制限を受けない
- 重い処理にも対応可能
- バッチ処理との相性が非常に良い
- 将来的な機能追加もしやすい
- 構成が比較的シンプル

### デメリット

- Lambda より起動時間は長い
- ECS Task Definition など管理対象が増える
- Fargate 実行料金が発生する

---

## 案3：Scheduler → Lambda → S3 → Lambda → SES

### 構成

```
EventBridge Scheduler
        │
        ▼
Lambda①（Private）
        │
        ├── RDS
        │
        ▼
S3
        │
   S3 Event
        ▼
Lambda②（Public）
        │
        ▼
SES
```

Lambda① は RDS から取得したメール本文を生成し、S3 へ保存する。

S3 イベントを契機に Lambda② を起動し、S3 から本文を取得して SES で送信する。

Lambda① は S3 Gateway Endpoint を利用することで、NAT Gateway なしで S3 へ保存可能。

Lambda② は VPC 外で実行することで SES へ直接接続可能。

### メリット

- NAT Gateway 不要
- メール本文を S3 へ保存できる
- 本文生成とメール送信を分離できる
- 再送処理を実装しやすい
- 障害時の調査がしやすい
- 将来的な拡張性が高い

### デメリット

- 構成が最も複雑
- Lambda が 2 つ必要
- S3 イベント管理が必要
- 障害ポイントが増える
- 保守コストが高い
- メール送信だけを考えるとオーバースペック

---

## 比較表

| 項目             | 案1    | 案2    | 案3    |
| -------------- | ----- | ----- | ----- |
| シンプルさ          | ★★★★★ | ★★★★☆ | ★★☆☆☆ |
| 運用しやすさ         | ★★★★★ | ★★★★☆ | ★★☆☆☆ |
| 拡張性            | ★★★☆☆ | ★★★★★ | ★★★★★ |
| 重い処理への対応       | ★★☆☆☆ | ★★★★★ | ★★★★☆ |
| NAT 不要         | ×     | ○     | ○     |
| 保守性            | ★★★★★ | ★★★★☆ | ★★☆☆☆ |
| 障害切り分け         | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| メール再送          | △     | △     | ★★★★★ |

---

## 採用案

### 採用

**案2（Scheduler → ECS RunTask → SES）**

### 採用理由

今回の要件では以下を満たす必要がある。

- RDS からデータ取得
- メール本文生成
- SES でメール送信
- NAT Gateway を構築しない
- 保守性を高く保ちたい

案1 では Lambda を VPC 内へ配置する必要があり、SES へ接続するために NAT Gateway または SES VPC Endpoint が必要となる。今回の方針では NAT Gateway を採用しないため、追加のネットワーク構成が必要となる。

案3 では NAT Gateway を利用せず実現可能であるが、Lambda を 2 つに分割し、S3 を経由するため構成が複雑になる。メール本文の永続化や再送機能が必須要件であれば有力な選択肢となるが、今回の要件ではそこまでの構成は不要と判断した。

一方、案2 では RunTask を Public Subnet へ配置し、Public IP を付与することで SES へ直接アクセスできる。また、RDS は同一 VPC 内通信となるため Private Subnet に配置された RDS へ安全に接続できる。

この構成であれば

- NAT Gateway 不要
- 構成がシンプル
- Go の既存資産をそのまま利用可能
- 将来的なバッチ追加にも対応しやすい
- Lambda の制限を考慮する必要がない

というメリットがあり、今回の要件を最もシンプルかつ保守しやすい形で満たせると判断した。

---

## 結論

今回のシステム要件を考慮した結果、

**「EventBridge Scheduler → ECS RunTask（Public Subnet）→ RDS → SES」**

を採用する。

NAT Gateway を構築することなく実装でき、構成のシンプルさ、保守性、拡張性のバランスが最も良い構成である。
