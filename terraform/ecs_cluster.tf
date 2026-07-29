resource "aws_ecs_cluster" "app" {
  name = "${var.app_name}-cluster"

  setting {
    name  = "containerInsights"
    value = "disabled"
  }
}

resource "aws_cloudwatch_log_group" "gmail_login" {
  name              = "/ecs/${var.app_name}/gmail-login"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "check_emails" {
  name              = "/ecs/${var.app_name}/check-emails"
  retention_in_days = 30
}
