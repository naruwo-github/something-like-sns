
# ECSクラスタ
resource "aws_ecs_cluster" "main" {
  name = "sns-app-cluster"
}

# ECSサービス用のセキュリティグループ
resource "aws_security_group" "ecs_service" {
  name        = "sns-app-ecs-service-sg"
  description = "Security group for the ECS service"
  vpc_id      = aws_vpc.main.id

  # ALBからのインバウンドトラフィックを許可
  ingress {
    from_port       = 0 # 動的ポートマッピングを考慮してすべてのポートを許可
    to_port         = 0
    protocol        = "-1"
    security_groups = [aws_security_group.alb.id]
  }

  # アウトバウンドはすべて許可
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# --- APIタスク定義 ---
resource "aws_ecs_task_definition" "api" {
  family                   = "sns-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"  # 0.25 vCPU
  memory                   = "512"  # 512MB
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  # task_role_arn            = aws_iam_role.ecs_task_execution_role.arn # 必要に応じてタスク用の別ロールを指定

  tags = {
    "Name" = "sns-api-task"
  }

  container_definitions = jsonencode([
    {
      name      = "sns-api-container"
      image     = "${aws_ecr_repository.api.repository_url}:latest"
      essential = true
      portMappings = [
        {
          containerPort = 8080,
          hostPort      = 8080
        }
      ]
      # Secrets ManagerからDB接続情報を環境変数として取得
      secrets = [
        {
          name      = "DB_USER",
          valueFrom = "${aws_secretsmanager_secret.db_credentials.arn}:username::"
        },
        {
          name      = "DB_PASS",
          valueFrom = "${aws_secretsmanager_secret.db_credentials.arn}:password::"
        }
      ],
      environment = [
        {
          name  = "DB_HOST",
          value = aws_db_instance.main.address
        },
        {
          name  = "DB_PORT",
          value = tostring(aws_db_instance.main.port)
        },
        {
          name  = "DB_NAME",
          value = replace(var.project_name, "-", "")
        }
      ],
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.api_logs.name,
          "awslogs-region"        = "ap-northeast-1",
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

# --- DBマイグレーションタスク定義 ---

resource "aws_ecs_task_definition" "db_migration" {
  family                   = "sns-db-migration"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name      = "sns-db-migration-container"
      image     = "${aws_ecr_repository.db_migration.repository_url}:latest"
      essential = true
      # migrateコマンドは環境変数でDB接続情報を受け取る
      environment = [
        {
            # ex) mysql://user:pass@tcp(host:port)/dbname
            name = "DATABASE_URL",
            value = "mysql://${jsondecode(data.aws_secretsmanager_secret_version.db_credentials.secret_string).username}:${jsondecode(data.aws_secretsmanager_secret_version.db_credentials.secret_string).password}@tcp(${aws_db_instance.main.address}:${aws_db_instance.main.port})/${replace(var.project_name, "-", "" )}"
        }
      ],
      command = [
        "-path",
        "/migrations",
        "-database",
        "$(DATABASE_URL)",
        "up"
      ],
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.db_migration_logs.name,
          "awslogs-region"        = "ap-northeast-1",
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

# CloudWatch Logs グループ
resource "aws_cloudwatch_log_group" "api_logs" {
  name              = "/ecs/sns-api"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "db_migration_logs" {
  name              = "/ecs/sns-db-migration"
  retention_in_days = 7
}

# --- ECSサービス ---
resource "aws_ecs_service" "main" {
  name            = "sns-api-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  launch_type     = "FARGATE"
  desired_count   = 1 # ひとまず1つで起動

  network_configuration {
    subnets         = [aws_subnet.private[0].id, aws_subnet.private[1].id]
    security_groups = [aws_security_group.ecs_service.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "sns-api-container"
    container_port   = 8080
  }

  # ALBのヘルスチェックを待つために必要
  depends_on = [aws_lb_listener.http]
}
