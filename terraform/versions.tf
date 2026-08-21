terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # for non-DA deployment, in account without existing infra bucket
  # command, in terraform/boostrap: tf init && tf apply
  #
  # Bucket/table created by terraform/bootstrap — run that first. Backend
  # blocks can't reference variables, so these are literal and must match
  # bootstrap's state_bucket_name/lock_table_name if you change those.
  # backend "s3" {
  #   bucket         = "go-email-agent-terraform-state"
  #   key            = "go-email-agent/terraform.tfstate"
  #   region         = "us-west-1"
  #   use_lockfile   = true
  #   encrypt        = true
  # }

  # for DA deployment, using existing infra bucket
  # command:  terraform init -backend-config=backend.dev.hcl
  backend "s3" {
    encrypt      = true
    use_lockfile = true
  }
}
