resource "aws_security_group" "gmail_login_task" {
  name        = "${var.app_name}-gmail-login-task"
  description = "gmail_login ECS task - inbound only from the ALB"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_ecs_task_definition" "gmail_login" {
  family                   = "${var.app_name}-gmail-login"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.gmail_login_task.arn

  container_definitions = jsonencode([
    {
      name    = "gmail-login"
      image   = "${aws_ecr_repository.app.repository_url}:${var.image_tag}"
      command = ["/usr/local/bin/gmail_login"]

      portMappings = [{
        containerPort = 8080
        protocol      = "tcp"
      }]

      # Full shared config set (see variables.tf) even though this binary
      # only actually reads a subset of it — one consistent env block for
      # both services, per the "keep things simple" ask.
      environment = [
        { name = "GOOGLE_CLIENT_ID", value = var.google_client_id },
        { name = "AWS_REGION", value = var.aws_region },
        { name = "ANTHROPIC_MODEL_ID", value = var.anthropic_model_id },
        { name = "GOOGLE_OAUTH_REDIRECT_URL", value = "https://${var.domain_name}/callback" },
        { name = "S3_BUCKET", value = var.s3_bucket_name },
        { name = "SSM_PREFIX", value = var.ssm_prefix },
        # Deliberately no STORAGE var here — leaving it unset (rather than
        # "local") means the app uses the real SSM/S3-backed storage. Setting
        # it to "local" in ECS would silently write tokens/agent files to the
        # container's ephemeral filesystem and lose them on every restart.
      ]

      secrets = [
        { name = "GOOGLE_CLIENT_SECRET", valueFrom = aws_ssm_parameter.google_client_secret.arn },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.gmail_login.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "gmail-login"
        }
      }
    }
  ])
}

resource "aws_ecs_service" "gmail_login" {
  name            = "${var.app_name}-gmail-login"
  cluster         = aws_ecs_cluster.app.id
  task_definition = aws_ecs_task_definition.gmail_login.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.gmail_login_task.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.gmail_login.arn
    container_name   = "gmail-login"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener.gmail_login_https]
}
