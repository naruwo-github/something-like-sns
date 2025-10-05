# IAM Role for CodePipeline
resource "aws_iam_role" "codepipeline_role" {
  name = "${var.project_name}-codepipeline-role"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = {
          Service = "codepipeline.amazonaws.com"
        }
        Action    = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_policy" "codepipeline_policy" {
  name        = "${var.project_name}-codepipeline-policy"
  description = "Policy for CodePipeline to access artifacts, trigger builds and deployments."

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow",
        Action   = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:GetBucketVersioning",
          "s3:PutObjectAcl",
          "s3:PutObject"
        ],
        Resource = [
          aws_s3_bucket.codepipeline_artifacts.arn,
          "${aws_s3_bucket.codepipeline_artifacts.arn}/*"
        ]
      },
      {
        Effect   = "Allow",
        Action   = "codestar-connections:UseConnection",
        Resource = var.codestar_connection_arn
      },
      {
        Effect   = "Allow",
        Action   = [
          "codebuild:StartBuild",
          "codebuild:BatchGetBuilds"
        ],
        Resource = [
          aws_codebuild_project.main.arn,
          aws_codebuild_project.run_migration.arn
        ]
      },
      {
        Effect   = "Allow",
        Action   = [
          "ecs:DescribeServices",
          "ecs:DescribeTaskDefinition",
          "ecs:DescribeTasks",
          "ecs:ListTasks",
          "ecs:RegisterTaskDefinition",
          "ecs:UpdateService"
        ],
        Resource = [
          aws_ecs_service.main.id,
          aws_ecs_task_definition.api.arn,
          aws_ecs_task_definition.db_migration.arn,
          aws_ecs_cluster.main.arn
        ]
      },
      {
        Effect   = "Allow",
        Action   = "ecs:RunTask",
        Resource = aws_ecs_task_definition.db_migration.arn,
        Condition = {
          StringEquals = {
            "ecs:cluster" = aws_ecs_cluster.main.arn
          }
        }
      },
      {
        Effect   = "Allow",
        Action   = "iam:PassRole",
        Resource = aws_iam_role.ecs_task_execution_role.arn,
        Condition = {
          StringEqualsIfExists = {
            "iam:PassedToService" = "ecs-tasks.amazonaws.com"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "codepipeline_attach" {
  role       = aws_iam_role.codepipeline_role.name
  policy_arn = aws_iam_policy.codepipeline_policy.arn
}

# S3 Bucket for Pipeline Artifacts
resource "aws_s3_bucket" "codepipeline_artifacts" {
  bucket = "${var.project_name}-codepipeline-artifacts-${data.aws_caller_identity.current.account_id}"
}

# CodeBuild Project
resource "aws_codebuild_project" "main" {
  name          = "${var.project_name}-build"
  description   = "Builds the Docker images for the application"
  service_role  = aws_iam_role.codebuild_role.arn
  build_timeout = "20" # minutes

  artifacts {
    type = "CODEPIPELINE"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:7.0"
    type                        = "LINUX_CONTAINER"
    privileged_mode             = true

    environment_variable {
      name  = "AWS_ACCOUNT_ID"
      value = data.aws_caller_identity.current.account_id
    }
    environment_variable {
      name  = "API_ECR_URI"
      value = aws_ecr_repository.api.repository_url
    }
    environment_variable {
      name  = "MIGRATE_ECR_URI"
      value = aws_ecr_repository.db_migration.repository_url
    }
    environment_variable {
      name  = "API_CONTAINER_NAME"
      value = var.api_container_name
    }
  }

  logs_config {
    cloudwatch_logs {
      group_name  = "/aws/codebuild/${var.project_name}"
      stream_name = "build"
    }
  }

  source {
    type            = "CODEPIPELINE"
    buildspec       = file("${path.module}/buildspec.yml")
  }
}

# CodePipeline
resource "aws_codepipeline" "main" {
  name     = "${var.project_name}-pipeline"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    type     = "S3"
    location = aws_s3_bucket.codepipeline_artifacts.id
  }

  stage {
    name = "Source"
    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["SourceOutput"]

      configuration = {
        ConnectionArn    = var.codestar_connection_arn
        FullRepositoryId = "${var.github_repo_owner}/${var.github_repo_name}"
        BranchName       = var.github_repo_branch
      }
    }
  }

  stage {
    name = "Build"
    action {
      name             = "Build"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      version          = "1"
      input_artifacts  = ["SourceOutput"]
      output_artifacts = ["BuildOutput"]

      configuration = {
        ProjectName = aws_codebuild_project.main.name
      }
    }
  }

  stage {
    name = "Deploy"
    
    action {
      name            = "DeployDBMigration"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      version         = "1"
      run_order       = 1
      input_artifacts = ["SourceOutput"]

      configuration = {
        ProjectName = aws_codebuild_project.run_migration.name
      }
    }

    action {
      name            = "DeployAPI"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "ECS"
      version         = "1"
      run_order       = 2
      input_artifacts = ["BuildOutput"]

      configuration = {
        ClusterName = aws_ecs_cluster.main.name
        ServiceName = aws_ecs_service.main.name
        FileName    = "imagedefinitions.json"
      }
    }
  }
}

# CodeBuild project to run DB migrations
resource "aws_codebuild_project" "run_migration" {
  name          = "${var.project_name}-run-migration"
  description   = "Runs database migrations via ECS Run Task"
  service_role  = aws_iam_role.codebuild_role.arn
  build_timeout = "10" # minutes

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:7.0"
    type                        = "LINUX_CONTAINER"
  }

  source {
    type      = "NO_SOURCE"
    buildspec = <<-EOT
      version: 0.2
      phases:
        build:
          commands:
            - |
              NETWORK_CONFIG=$(cat <<EOF
              {
                "awsvpcConfiguration": {
                  "subnets": [
                    "${join("\",\"", aws_subnet.private.*.id)}"
                  ],
                  "securityGroups": [
                    "${aws_security_group.ecs_service.id}"
                  ],
                  "assignPublicIp": "DISABLED"
                }
              }
              EOF
              )
              aws ecs run-task --cluster ${aws_ecs_cluster.main.name} --task-definition ${aws_ecs_task_definition.db_migration.arn} --launch-type FARGATE --network-configuration "$NETWORK_CONFIG"
    EOT
  }
}
