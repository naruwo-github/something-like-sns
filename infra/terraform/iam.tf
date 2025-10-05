#  AWSの各サービスが他のサービスを操作するために必要な「ロール（役割）」という権限を定義します。
#   * aws_iam_role (計2個): ECSタスク実行用と、将来使うCodeBuild用の2つのロールを作成します。
#   * aws_iam_role_policy_attachment (1個): ECSタスク実行ロールに、AWSが管理する標準的な権限ポリシーをアタッチ（紐付け）します。

#  【このファイルの合計: 3リソース】

# ECSタスクがECRからイメージをプルしたり、CloudWatch Logsにログを書き込むためのIAMロール
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "${var.project_name}-ecs-task-execution-role"

  # このロールをECSタスクが引き受ける(AssumeRole)ための信頼ポリシー
  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action    = "sts:AssumeRole",
        Effect    = "Allow",
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-ecs-task-execution-role"
  }
}

# 上記のロールにAWS管理ポリシー「AmazonECSTaskExecutionRolePolicy」をアタッチ
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# CodeBuildプロジェクトが使用するIAMロール
# NOTE: これは雛形です。フェーズ4でCI/CDパイプラインを構築する際に、より具体的な権限を追加します。
resource "aws_iam_role" "codebuild_role" {
  name = "${var.project_name}-codebuild-role"

  # このロールをCodeBuildサービスが引き受けるための信頼ポリシー
  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action    = "sts:AssumeRole",
        Effect    = "Allow",
        Principal = {
          Service = "codebuild.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-codebuild-role"
  }
}

resource "aws_iam_policy" "codebuild_policy" {
  name        = "${var.project_name}-codebuild-policy"
  description = "Policy for the main CodeBuild project"

  policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:GetBucketVersioning",
          "s3:PutObject",
          "s3:PutObjectAcl"
        ],
        Resource = [
          aws_s3_bucket.codepipeline_artifacts.arn,
          "${aws_s3_bucket.codepipeline_artifacts.arn}/*"
        ]
      },
      {
        Effect   = "Allow",
        Action   = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource = "*"
      },
      {
        Effect   = "Allow",
        Action   = "ecr:GetAuthorizationToken",
        Resource = "*"
      },
      {
        Effect   = "Allow",
        Action   = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage"
        ],
        Resource = [
          aws_ecr_repository.api.arn,
          aws_ecr_repository.db_migration.arn
        ]
      },
      {
        Effect   = "Allow",
        Action   = "secretsmanager:GetSecretValue",
        Resource = aws_secretsmanager_secret.db_credentials.arn
      },
      {
        Effect   = "Allow",
        Action   = "ecs:RunTask",
        Resource = aws_ecs_task_definition.db_migration.arn
      },
      {
        Effect   = "Allow",
        Action   = "iam:PassRole",
        Resource = aws_iam_role.ecs_task_execution_role.arn
      },
      {
        Effect   = "Allow",
        Action   = "ssm:GetParameters",
        Resource = "arn:aws:ssm:ap-northeast-1:291945306023:parameter/CodeBuild/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "codebuild_attach" {
  role       = aws_iam_role.codebuild_role.name
  policy_arn = aws_iam_policy.codebuild_policy.arn
}


resource "aws_iam_role_policy" "ecs_secrets_inline_policy" {
  name = "ecs-secrets-inline-policy"
  role = aws_iam_role.ecs_task_execution_role.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = "secretsmanager:GetSecretValue",
        Effect   = "Allow",
        Resource = "arn:aws:secretsmanager:ap-northeast-1:291945306023:secret:sns-app/db-credentials-*"
      }
    ]
  })
}
