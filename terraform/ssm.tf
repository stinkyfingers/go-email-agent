# App-level secrets, managed by Terraform. Per-user Gmail OAuth tokens are
# NOT here — those live dynamically under var.ssm_prefix, created at runtime
# by gmail_login as users authenticate (see iam.tf for the scoped access
# granted to that path instead of resources here).

resource "aws_ssm_parameter" "google_client_secret" {
  name  = "/${var.app_name}/google-client-secret"
  type  = "SecureString"
  value = var.google_client_secret
}

# SSM parameter values can't be empty strings, and an empty webhook means
# "Slack notifications disabled" (see llm/slack.go) — so when unset, don't
# create the parameter at all rather than storing a placeholder that the app
# would actually try to POST to.
resource "aws_ssm_parameter" "slack_webhook" {
  count = var.slack_webhook != "" ? 1 : 0
  name  = "/${var.app_name}/slack-webhook"
  type  = "SecureString"
  value = var.slack_webhook
}


resource "aws_ssm_parameter" "hexagon_bearer_token" {
  name  = "/${var.app_name}/hexagon_bearer_token"
  type  = "SecureString"
  value = var.hexagon_bearer_token
}