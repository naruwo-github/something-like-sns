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
