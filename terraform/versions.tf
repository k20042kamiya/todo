terraform {
  required_version = ">= 1.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

}

provider "aws" {
  region = var.aws_region
}

# CloudFront用ACMはus-east-1が必須
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}
