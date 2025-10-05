# AWSリージョン
variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

# プロジェクト名 (リソースのプレフィックスとして使用)
variable "project_name" {
  description = "Project name for resource prefix"
  type        = string
  default     = "sns-app"
}

# VPCのCIDRブロック
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# RDSデータベースのマスターユーザー名
variable "db_master_username" {
  description = "Master username for the RDS database"
  type        = string
  default     = "admin"
}

# APIコンテナ名
variable "api_container_name" {
  description = "API container name"
  type        = string
  default     = "sns-api-container"
}

# GitHubリポジトリの所有者名
variable "github_repo_owner" {
  type        = string
  description = "GitHub repository owner (e.g., 'your-username')"
}

# GitHubリポジトリ名
variable "github_repo_name" {
  type        = string
  description = "GitHub repository name (e.g., 'something-like-sns')"
}

# CodePipelineが追跡するブランチ名
variable "github_repo_branch" {
  type        = string
  default     = "main"
  description = "GitHub repository branch to track"
}

# CodeStar ConnectionのARN
variable "codestar_connection_arn" {
  type        = string
  description = "ARN of the CodeStar connection to GitHub"
}
