> **このドキュメントの目的**: このアプリケーションをAWSクラウドにデプロイするためのTODOリストです。

# AWSインフラ構築・デプロイ計画 TODO (効率化版)

`SYSTEM_ARCHITECTURE.md`の構成を実現するためのタスクリスト。インフラ定義(Terraform)とデプロイ・動作確認をコンポーネント単位でまとめ、手戻りを減らし手際よく進める。

---

## フェーズ1: ネットワーク・DB・コンテナレジストリの基盤構築 (Terraform)

アプリケーションを配置する前の、共有基盤となるネットワーク、データベース、コンテナレジストリをまとめて構築する。

- [x] **Terraform初期設定**:
    - [x] `infra/terraform` ディレクトリとAWSプロバイダを設定する
    - [x] state管理用のS3バケットとDynamoDBテーブルを作成する
- [x] **VPC構築**:
    - [x] VPC、パブリック/プライベートサブネット、インターネットゲートウェイ等を定義する
- [x] **IAMロール準備**:
    - [x] ECSタスク実行ロール、CodeBuild用ロールの雛形を作成する
- [x] **ECRリポジトリ作成**:
    - [x] API用とDBマイグレーション用のECRリポジトリを2つ作成する
- [x] **Secrets Manager設定**:
    - [x] DBのマスターパスワードを管理するシークレットを作成する
- [x] **RDS for MySQL構築**:
    - [x] プライベートサブネットにRDS for MySQLを構築し、Secrets Managerの値を参照させる

---

## フェーズ2: バックエンドの完全なデプロイと動作確認 (Terraform & 手動)

バックエンドが単体で動作するところまでを、Terraformと手動作業を組み合わせて一気に行う。

- [x] **ALBとECSクラスタ構築 (Terraform)**:
    - [x] ALBをパブリックサブネットに作成する
    - [x] ECSクラスタを作成する
- [x] **ECSタスク定義の作成 (Terraform)**:
    - [x] API用のタスク定義を作成する (ECRイメージURI、Secrets Manager参照を含む)
    - [x] DBマイグレーション用のタスク定義を作成する
- [x] **Dockerイメージのビルドとプッシュ (ローカルPC)**:
    - [x] APIのDockerイメージをビルドし、ECRにプッシュする
    - [x] DBマイグレーションのDockerイメージをビルドし、ECRにプッシュする
- [x] **デプロイと動作確認 (手動)**:
    - [x] ECS Run TaskでDBマイグレーションタスクを実行する
    - [x] ECSサービスをTerraformで作成し、APIタスクを起動する
    - [x] ALB経由でAPIのヘルスチェックや簡単なエンドポイントにアクセスし、DB接続を含めて正常動作を確認する

### フェーズ2完了時の到達点（稼働中コンポーネント）

- **ネットワーク(VPC)**
    - **VPC**: `sns-app-vpc`（10.0.0.0/16）
    - **サブネット**: パブリック×2（ALB用）、プライベート×2（ECS/RDS用）
    - **IGW / NAT / ルートテーブル**: 構成済み（プライベート→NAT経由で外部到達可）
- **セキュリティグループ**
    - **ALB**: `sns-app-alb-sg`（0.0.0.0/0 → 80許可）
    - **ECSサービス**: `sns-app-ecs-service-sg`（ALB SG からのインバウンド許可）
    - **RDS**: `${project_name}-rds-sg`（VPC内 3306 許可）
- **ロードバランサ**
    - **ALB**: `sns-app-alb`（HTTP:80）
    - **ターゲットグループ**: `sns-app-api-tg`（ターゲットタイプ: ip, ポート: 8080, ヘルスチェック: `/health`）
    - **DNS**: `sns-app-alb-794081207.ap-northeast-1.elb.amazonaws.com`
    - **状態**: `/health` 200 OK 確認済み
- **コンテナ実行基盤**
    - **ECS クラスタ**: `sns-app-cluster`
    - **ECS サービス**: `sns-api-service`（Fargate, desired=1, プライベートサブネット配置）
    - **タスク定義(API)**: `sns-api`
        - コンテナ: `sns-api-container`（ポート 8080, awslogs `/ecs/sns-api`）
        - 環境変数: `DB_HOST`/`DB_PORT`/`DB_NAME`（Terraform 供給）
        - シークレット: `DB_USER`/`DB_PASS`（Secrets Manager 参照）
- **データベース**
    - **RDS MySQL 8.0**: `sns-app-db`（db.t3.micro, プライベート）
    - **DB名**: `snsapp`
    - **状態**: API `/dbping` にて `{"status":"up"}` 確認
- **マイグレーション**
    - **ECRイメージ**: `sns-app/db-migration`
    - **タスク定義**: `sns-db-migration`
    - **実行結果**: RunTask 成功（exitCode=0、`1/u init_schema` 適用済み）
    - **ログ**: `/ecs/sns-db-migration`
- **コンテナレジストリ(ECR)**
    - `sns-app/api`（APIイメージ）
    - `sns-app/db-migration`（マイグレーションイメージ）
- **IAM / シークレット**
    - **ロール**: `sns-app-ecs-task-execution-role`（ECR pull, CloudWatch Logs, Secrets 参照権限付与）
    - **Secrets Manager**: `sns-app/db-credentials`（`username`, `password`）
- **可観測性**
    - **CloudWatch Logs**: `/ecs/sns-api`, `/ecs/sns-db-migration` 稼働中

---

## フェーズ3: フロントエンドのデプロイとE2E動作確認 (Vercel)

動作しているバックエンドに接続するフロントエンドをVercelにデプロイし、アプリケーション全体の疎通を確認する。

- [x] **Vercel プロジェクト作成**:
    - [x] `apps/web` を対象にリンク/作成する
    - [x] 環境変数 `NEXT_PUBLIC_API_BASE` にALBのURLを設定する
- [x] **デプロイと動作確認**:
    - [x] `vercel deploy --prod` で本番デプロイ
    - [x] 発行URLで表示・投稿・リロード・詳細/コメントまで確認する

---

## フェーズ4: CI/CDパイプラインの構築と完全自動化 (バックエンドのみTerraform)

バックエンド側の自動化を行う（フロントはVercelのGit連携で自動デプロイ）。

- [x] **CodePipelineとCodeBuildの作成 (Terraform)**:
    - [x] CodePipelineとCodeBuildのプロジェクトを定義する
    - [x] `buildspec.yml` にビルドからデプロイアーティファクト作成までの一連のコマンドを記述する
- [x] **デプロイフローの自動化 (Terraform)**:
    - [x] パイプラインにSource(Git), Build(CodeBuild)ステージを定義する
    - [x] Deployステージを定義する:
        1.  DBマイグレーション (ECS Run Task)
        2.  バックエンドデプロイ (ECSサービス更新)
        3.  フロントエンドはVercelにより自動デプロイ（本パイプライン対象外）
- [x] **自動化の最終テスト**:
    - [x] Gitリポジトリへのプッシュをトリガーに、パイプライン全体が正常に動作し、アプリケーションが更新されることを確認する
