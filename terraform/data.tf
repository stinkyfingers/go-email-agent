data "aws_vpc" "default" {
  id = var.vpc_id
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }

  # This VPC has both a public and private subnet per AZ (see
  # public-subnet-<az>/private-subnet-<az> Name tags). An ALB requires at
  # most one subnet per AZ, and — since this ALB is internet-facing and ECS
  # tasks rely on assign_public_ip rather than a NAT gateway — it needs to be
  # the public ones specifically (verified via route table: these route
  # 0.0.0.0/0 to a real Internet Gateway, unlike the private pair).
  filter {
    name   = "tag:Name"
    values = ["public-*"]
  }
}

data "aws_caller_identity" "current" {}
