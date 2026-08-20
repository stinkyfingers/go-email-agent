resource "aws_security_group" "check_emails_task" {
  name        = "${var.app_name}-check-emails-task"
  description = "check_emails ECS task — outbound only, never receives inbound traffic"
  vpc_id      = data.aws_vpc.default.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

locals {
  # GOOGLE_CLIENT_SECRET is always required; SLACK_WEBHOOK's SSM parameter
  # only exists when var.slack_webhook was actually set (see ssm.tf), so
  # it's included conditionally rather than referencing a resource that may
  # not have been created.
  check_emails_secrets = concat(
    [{ name = "GOOGLE_CLIENT_SECRET", valueFrom = aws_ssm_parameter.google_client_secret.arn }],
    length(aws_ssm_parameter.slack_webhook) > 0 ? [{ name = "SLACK_WEBHOOK", valueFrom = aws_ssm_parameter.slack_webhook[0].arn }] : []
  )
}

resource "aws_ecs_task_definition" "check_emails" {
  family                   = "${var.app_name}-check-emails"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "1024"
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.check_emails_task.arn

  container_definitions = jsonencode([
    {
      name    = "check-emails"
      image   = "${aws_ecr_repository.app.repository_url}:${var.image_tag}"
      command = ["/usr/local/bin/check_emails"]

      environment = [
        { name = "AWS_REGION", value = var.aws_region },
        { name = "ANTHROPIC_MODEL_ID", value = var.anthropic_model_id },
        { name = "HEXAGON_URL", value = var.hexagon_url },
        { name = "HEXAGON_BEARER_TOKEN", value = var.hexagon_bearer_token },
        { name = "GOOGLE_CLIENT_ID", value = var.google_client_id },
        { name = "GOOGLE_OAUTH_REDIRECT_URL", value = "https://${var.domain_name}/callback" },
        { name = "S3_BUCKET", value = var.s3_bucket_name },
        { name = "SSM_PREFIX", value = var.ssm_prefix },
      ]

      secrets = local.check_emails_secrets

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.check_emails.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "check-emails"
        }
      }
    }
  ])
}

# Runs check_emails on a schedule via classic EventBridge (CloudWatch Events)
# rate rule -> ECS RunTask, rather than a long-lived service — matches the
# "bounded batch job" shape of this workload.
resource "aws_cloudwatch_event_rule" "check_emails_schedule" {
  name                = "${var.app_name}-check-emails-schedule"
  schedule_expression = "rate(15 minutes)"
}

resource "aws_cloudwatch_event_target" "check_emails" {
  rule     = aws_cloudwatch_event_rule.check_emails_schedule.name
  arn      = aws_ecs_cluster.app.arn
  role_arn = aws_iam_role.eventbridge_run_task.arn

  ecs_target {
    task_definition_arn = aws_ecs_task_definition.check_emails.arn
    task_count          = 1
    launch_type         = "FARGATE"

    network_configuration {
      subnets          = data.aws_subnets.default.ids
      security_groups  = [aws_security_group.check_emails_task.id]
      assign_public_ip = true
    }
  }
}
