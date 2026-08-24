data "aws_iam_policy_document" "ecs_tasks_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# --- Execution role: shared by both services. Used by the ECS agent itself
# to pull the image and resolve the "secrets" block into env vars — NOT used
# by application code. ---

resource "aws_iam_role" "execution" {
  name               = "${var.app_name}-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume.json
}

resource "aws_iam_role_policy_attachment" "execution_managed" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "execution_secrets" {
  statement {
    # The ECS agent calls the plural ssm:GetParameters (not GetParameter) to
    # resolve the "secrets" block — granting only the singular action is a
    # common gotcha that silently fails container startup.
    actions = ["ssm:GetParameters"]
    resources = compact([
      aws_ssm_parameter.google_client_secret.arn,
      try(aws_ssm_parameter.slack_webhook[0].arn, ""),
      aws_ssm_parameter.hexagon_bearer_token.arn,
    ])
  }
}

resource "aws_iam_role_policy" "execution_secrets" {
  name   = "${var.app_name}-execution-secrets"
  role   = aws_iam_role.execution.id
  policy = data.aws_iam_policy_document.execution_secrets.json
}

# --- gmail_login task role: only ever creates/removes tokens (OAuth
# callback and logout), never reads them or touches S3/Bedrock/Hexagon —
# kept minimal. ---

resource "aws_iam_role" "gmail_login_task" {
  name               = "${var.app_name}-gmail-login-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume.json
}

data "aws_iam_policy_document" "gmail_login_task" {
  statement {
    actions = [
      "ssm:PutParameter",
      "ssm:DeleteParameter",
    ]
    resources = ["arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter${var.ssm_prefix}/*"]
  }
}

resource "aws_iam_role_policy" "gmail_login_task" {
  name   = "${var.app_name}-gmail-login-task"
  role   = aws_iam_role.gmail_login_task.id
  policy = data.aws_iam_policy_document.gmail_login_task.json
}

# --- check_emails task role: reads (never writes) tokens, reads agent files
# from S3, and calls Bedrock. ---

resource "aws_iam_role" "check_emails_task" {
  name               = "${var.app_name}-check-emails-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume.json
}

data "aws_iam_policy_document" "check_emails_task" {
  statement {
    actions = [
      "ssm:GetParameter",
      "ssm:GetParametersByPath",
    ]
    resources = ["arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter${var.ssm_prefix}/*"]
  }

  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.agentfiles.arn}/*"]
  }

  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.agentfiles.arn]
  }

  statement {
    actions = [
      "bedrock:InvokeModel",
      "bedrock:InvokeModelWithResponseStream",
    ]
    resources = [
      "arn:aws:bedrock:*::foundation-model/anthropic.claude-*",
      "arn:aws:bedrock:${var.aws_region}:${data.aws_caller_identity.current.account_id}:inference-profile/*",
    ]
  }
}

resource "aws_iam_role_policy" "check_emails_task" {
  name   = "${var.app_name}-check-emails-task"
  role   = aws_iam_role.check_emails_task.id
  policy = data.aws_iam_policy_document.check_emails_task.json
}

# --- EventBridge rule role: lets the scheduled rule invoke ECS RunTask for
# the check_emails task definition. ---

resource "aws_iam_role" "eventbridge_run_task" {
  name = "${var.app_name}-eventbridge-run-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "events.amazonaws.com" }
    }]
  })
}

data "aws_iam_policy_document" "eventbridge_run_task" {
  statement {
    actions = ["ecs:RunTask"]
    # Wildcarded on revision (":*") rather than referencing
    # aws_ecs_task_definition.check_emails.arn directly, since that ARN
    # pins one specific revision and would need a new IAM policy on every
    # task def update.
    resources = ["arn:aws:ecs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:task-definition/${var.app_name}-check-emails:*"]
    condition {
      test     = "ArnLike"
      variable = "ecs:cluster"
      values   = [aws_ecs_cluster.app.arn]
    }
  }

  statement {
    actions = ["iam:PassRole"]
    resources = [
      aws_iam_role.execution.arn,
      aws_iam_role.check_emails_task.arn,
    ]
  }
}

resource "aws_iam_role_policy" "eventbridge_run_task" {
  name   = "${var.app_name}-eventbridge-run-task"
  role   = aws_iam_role.eventbridge_run_task.id
  policy = data.aws_iam_policy_document.eventbridge_run_task.json
}
