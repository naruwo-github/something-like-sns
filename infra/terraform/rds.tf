#  MySQLデータベース（RDS）とその周辺設定を定義します。
#   * aws_db_instance (1個): MySQLデータベース本体です。
#   * aws_db_subnet_group (1個): データベースをどのプライベートサブネットに配置するかを定義するグループです。
#   * aws_security_group (1個): データベースへのアクセスを制御するファイアウォール。VPC内からのMySQL通信のみを許可します。

#  【このファイルの合計: 3リソース】

# RDSインスタンス用のセキュリティグループ
resource "aws_security_group" "rds" {
  name        = "${var.project_name}-rds-sg"
  description = "Allow inbound traffic to RDS from within the VPC"
  vpc_id      = aws_vpc.main.id

  # VPC内部からのMySQLトラフィック(ポート3306)を許可
  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  # アウトバウンドはすべて許可
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-rds-sg"
  }
}

# RDSインスタンスを配置するサブネットグループ
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-rds-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "${var.project_name}-rds-subnet-group"
  }
}

# Secrets Managerから最新のシークレット値を取得するためのデータソース
data "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
}

# RDS for MySQLインスタンス本体
resource "aws_db_instance" "main" {
  identifier             = "${var.project_name}-db"
  allocated_storage      = 20
  storage_type           = "gp2"
  engine                 = "mysql"
  engine_version         = "8.0"
  instance_class         = "db.t3.micro" # 開発用に小規模なインスタンスタイプを選択
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  db_name                = replace(var.project_name, "-", "")

  # Secrets Managerから取得した認証情報を使用
  username = jsondecode(data.aws_secretsmanager_secret_version.db_credentials.secret_string)["username"]
  password = jsondecode(data.aws_secretsmanager_secret_version.db_credentials.secret_string)["password"]

  # 開発用途のため、Multi-AZ配置と最終スナップショット取得は無効化
  multi_az               = false
  skip_final_snapshot    = true
  publicly_accessible    = false

  tags = {
    Name = "${var.project_name}-db-instance"
  }
}
