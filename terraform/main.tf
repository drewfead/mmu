terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4"
    }
    archive = {
      source = "hashicorp/archive"
    }
    null = {
      source = "hashicorp/null"
    }
  }

  backend "s3" {
    bucket         = "mmu-infra"
    key            = "terraform/tfstate"
    region         = "us-west-2"
    dynamodb_table = "mmu-tfstate"
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-west-2"
}