# Terraformと各種プロバイダのバージョン設定
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  # NOTE: 'terraform init'を実行する前に、Terraformの状態(tfstate)を保存するためのS3バケットとDynamoDBテーブルを手動で作成
  backend "s3" {
    bucket         = "sns-app-tfstate-ap-northeast-1"
    key            = "something-like-sns/phase1.tfstate"
    region         = "ap-northeast-1"
    dynamodb_table = "sns-app-terraform-locks"
    encrypt        = true
  }
}

# AWSプロバイダの設定 (使用するリージョン)
provider "aws" {
  region = var.region
}
