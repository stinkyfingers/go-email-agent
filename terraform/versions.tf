terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Bucket/table created by terraform/bootstrap — run that first. Backend
  # blocks can't reference variables, so these are literal and must match
  # bootstrap's state_bucket_name/lock_table_name if you change those.
  backend "s3" {
    bucket         = "go-email-agent-terraform-state"
    key            = "go-email-agent/terraform.tfstate"
    region         = "us-west-1"
    dynamodb_table = "go-email-agent-terraform-locks"
    encrypt        = true
  }
}
