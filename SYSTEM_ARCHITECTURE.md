# AWSシステム構成案

AWSクラウドにデプロイするためのシステム構成案

実際に使う抽象コンポーネント（外側→内側）
- クライアント（Web/Next.js）
- DNS（Route 53）
- CDN/エッジ（Amplify Hosting に内包）
- L7ロードバランサ（ALB）
- Web/SSRランタイム（Amplify Hosting 上のNext.js）
- アプリケーションサーバー（ECS Fargate のGo API）
- RDBMS（RDS for MySQL）
- CI/CD（CodePipeline / CodeBuild / ECR、マイグレーションはECS RunTask）
- 監視/可観測性（CloudWatch Logs/Metrics、X-Rayは検討）
- シークレット管理（Secrets Manager）
- ネットワーク（VPC/サブネット）

## システム構成図 (テキストベース)

```
                                      +-----------------------+
                                      |   AWS CodePipeline    |  <-- Gitリポジトリ (GitHubなど)
                                      +-----------+-----------+
                                                  | (CI/CD)
                                      +-----------v-----------+
                                      |    AWS CodeBuild      |
                                      | - pnpm install        |
                                      | - buf generate        |
                                      | - turbo build         |
                                      | - Docker build & push |
                                      | - Web artifact build  |
                                      +-----------------------+
                                                  |
                               +------------------+------------------+
                               |                  |                  |
         (フロントエンドデプロイ) |                  | (バックエンドデプロイ) |
+-----------------v----------+   +---------------v----------------+   +---------------v----------------+
|    AWS Amplify Hosting     |   |  Amazon ECR (Dockerレジストリ) |   |  ECSタスク (DBマイグレーション)    |
|  (CloudFrontは内包)        |   +--------------------------------+   +--------------------------------+
|  Next.js (SSR/SSG)         |
+----------------------------+
           ^
           | (HTTPS)
+----------+-----------------+
|   Amazon Route 53 (DNS)    |
+----------------------------+
           ^
           |
      [ エンドユーザー ]
           |
           | (APIコール)
           v
+----------------------------+
| Application Load Balancer  |
+----------------------------+
           |
           v
+----------------------------+
|   AWS Fargate on ECS       |
| +------------------------+ |
| | Go APIコンテナ         | |
| +------------------------+ |
+----------------------------+
           |
           v
+----------------------------+
|   Amazon RDS for MySQL     |
+----------------------------+

```

## コンポーネント詳細

### 1. フロントエンド (Next.js Web App)

* **AWS Amplify Hosting**
  * Next.jsのデプロイに最適化。CDN(CloudFront)はAmplify内で自動構成されるため、外部のCloudFrontは不要
  * 本プロジェクトではAmplify内蔵CIは使用せず、CodePipelineからアーティファクトを用いたデプロイAPIで反映する

### 2. バックエンド (Go API)

* **AWS Fargate on Amazon ECS**
  * コンテナ向けのサーバーレスコンピューティングエンジンで、サーバー管理なしにコンテナ化されたGoアプリケーションを実行するのに最適
  * 変動する負荷に応じてコンテナ数を自動的に調整するオートスケーリングをサポート
* **Amazon ECR (Elastic Container Registry)**
  * ビルドされたGo APIのDockerイメージを保存するためのプライベートコンテナレジストリ。CodeBuildがイメージをプッシュし、Fargateがサービスをデプロイするためにここからイメージをプル。
* **Application Load Balancer (ALB)**
  * 受信したAPIリクエストを受け付け、実行中の複数のFargateコンテナにトラフィックを分散する
  * SSL終端（HTTPS）やヘルスチェックも担当する

### 3. データベース

* **Amazon RDS for MySQL**
  * 高性能でスケーラブルなマネージドリレーショナルデータベース。開発環境がMySQLを使用しているため、互換性のあるRDS for MySQLは最適。
  * 当初はコストを抑えるため、**Multi-AZ配置は採用しない**。ただし、将来的な可用性向上のためのオプションとして検討可能とする。

### 4. CI/CD (継続的インテグレーション/継続的デプロイ)

* **AWS CodePipeline**
  * `git push`からデプロイまでを一元自動化するオーケストレーションサービス。Amplifyのデプロイもここに統合
* **AWS CodeBuild**
  * ソースのビルド、生成、コンテナイメージ作成、Webアーティファクト作成を実行：
    1. `pnpm install`
    2. `make proto`（または`buf generate`）
    3. `turbo build`（`api`/`web`）
    4. `api`の`docker build`とECRへのプッシュ
    5. `apps/web`のビルド成果物をzip化（Amplifyアーティファクト）
    6. CodePipelineの次ステージでAmplifyのManual Deploy APIによりリリース
* **データベースマイグレーション**
  * デプロイパイプライン内の専用ステージで、`packages/dbschema/migrations`からのスキーマ変更を適用する。
  * 安全なアプローチとして、新しいAPIコンテナをデプロイする**前**に、CodeBuildからECS RunTaskを呼び出して一度限りのマイグレーションタスク（`golang-migrate`）を実行する。

### 5. Infrastructure as Code (IaC)

* AWSリソースのプロビジョニングと管理は**Terraform**を用いて行う。
* インフラ構成をコード化することで、環境の再現性を担保し、手動オペレーションによるミスを削減する。

### 6. Observability (可観測性)

* **ロギング**: FargateコンテナやAmplifyアプリケーションのログは**Amazon CloudWatch Logs**に集約し、一元的に監視・検索できるようにする。
* **メトリクス**: 各AWSリソースのパフォーマンスメトリクス（CPU使用率、リクエスト数、レイテンシ等）は**Amazon CloudWatch Metrics**で収集・可視化し、アラートを設定する。
* **トレーシング**: APIリクエストのパフォーマンス分析やボトルネック特定のために**AWS X-Ray**の導入を検討する。

### 7. Secrets Management (シークレット管理)

* データベースのパスワードやAPIキーといった秘匿情報は**AWS Secrets Manager**を用いて安全に管理する。
* アプリケーションは実行時にIAMロールを通じてSecrets Managerから動的に認証情報を取得し、コードや環境変数に直接ハードコードすることを避ける。

### 8. ネットワーキングとその他のサービス

* **Amazon VPC (Virtual Private Cloud)**
  * AWSリソースのための隔離されたネットワーク空間。ALBは**パブリックサブネット**に、FargateコンテナとRDSデータベースは**プライベートサブネット**に配置する。これにより、データベースやバックエンドサービスへの直接的なインターネットアクセスを防ぎ、セキュリティを強化する。
* **Amazon Route 53**
  * カスタムドメイン（例: `example.com`）を管理するためのDNSサービスである。

## デプロイフロー概要

1. 開発者がGitリポジトリにコードをプッシュする。
2. **CodePipeline**がトリガーされる。
3. **CodeBuild**が`pnpm install`、`buf generate`、`turbo build`を実行し、APIのDockerイメージを**ECR**にプッシュ、Webのアーティファクト(zip)を作成。
4. （必要に応じて）ECS RunTaskでDBマイグレーションを実行する。
5. **CodePipeline**が**Fargate**サービスを更新（新しいECRイメージ）する。
6. **CodePipeline**が**Amplify**の手動デプロイAPI（CreateDeployment→zipアップロード→StartDeployment）を実行し、Webを更新する。