#  アプリケーションのDockerイメージを保存しておくためのプライベートな場所（ECR）を定義します。
#   * aws_ecr_repository (計2個): APIサーバー用と、DBマイグレーション用の2つのリポジトリを作成します。

#  【このファイルの合計: 2リソース】

# APIアプリケーションのDockerイメージを格納するECRリポジトリ
resource "aws_ecr_repository" "api" {
  name = "${var.project_name}/api"

  # タグを上書き可能にする
  image_tag_mutability = "MUTABLE"

  # プッシュ時にイメージをスキャンして脆弱性をチェックする
  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-api-repo"
  }
}

# データベースマイグレーションタスク用のDockerイメージを格納するECRリポジトリ
resource "aws_ecr_repository" "db_migration" {
  name = "${var.project_name}/db-migration"

  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-db-migration-repo"
  }
}
