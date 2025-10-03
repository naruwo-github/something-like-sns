# データベースのマスターユーザー用のランダムなパスワードを生成
resource "random_password" "db_master_password" {
  length  = 16
  special = true
  # URIやシェルで問題を起こしうる文字は避ける
  override_special = "_!%"
}

# DBの認証情報を格納するためのSecretをAWS Secrets Managerに作成
resource "aws_secretsmanager_secret" "db_credentials" {
  name = "${var.project_name}/db-credentials"
  tags = {
    Name = "${var.project_name}-db-credentials"
  }
}

# 上記で作成したSecretに、ユーザー名と生成したパスワードの初期値を設定
resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = var.db_master_username
    password = random_password.db_master_password.result
  })
}
