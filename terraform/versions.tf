terraform {
  required_version = ">= 1.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # 初回apply前にS3バケットとDynamoDBテーブルを手動作成してからコメントアウトを外す
  # backend "s3" {
  #   bucket         = "todo-app-terraform-state"
  #   key            = "prod/terraform.tfstate"
  #   region         = "ap-northeast-1"
  #   dynamodb_table = "todo-app-terraform-lock"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region
}

# CloudFront用ACMはus-east-1が必須
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}
