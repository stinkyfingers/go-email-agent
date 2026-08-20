variable "aws_region" {
  type    = string
  default = "us-west-1"
}

variable "app_name" {
  type    = string
  default = "go-email-agent"
}

variable "image_tag" {
  description = "Tag of the image in the ECR repo to deploy (e.g. a git SHA). No default on purpose — every deploy should pick one explicitly."
  type        = string
}

# --- Plain (non-secret) app config, injected as container env vars ---

variable "google_client_id" {
  type    = string
  default = "926019377623-lntcl176cqrichpkgh3gi6u6pe76na3k.apps.googleusercontent.com"
}

variable "anthropic_model_id" {
  type    = string
  default = "us.anthropic.claude-opus-4-8"
}

variable "hexagon_url" {
  description = <<-EOT
    URL of the Hexagon (ticketco-web) API. The
    "http://localhost:3001" default only works for local dev — inside ECS
    there is no ticketco-web on localhost, so this MUST be overridden to
    wherever that app is actually reachable from this VPC before the
    check_emails task can succeed at anything that uses Hexagon tools.
  EOT
  type        = string
  default     = "http://localhost:3001"
}

variable "hexagon_bearer_token" {
  description = "auth token for hexagon API"
  type        = string 
  sensitive   = true
}

variable "domain_name" {
  description = <<-EOT
    Public HTTPS domain for the gmail_login ALB (e.g. gmail-auth.example.com).
    Google requires HTTPS for OAuth redirect URIs, and ACM certs must be
    issued for a domain you control — there is no valid cert for the ALB's
    raw AWS DNS name. No default on purpose: this MUST be set, and the
    resulting URL (https://<domain_name>/callback) must also be added to
    this OAuth client's "Authorized redirect URIs" in Google Cloud Console
    (a manual, non-Terraform step) before login will work.
  EOT
  type        = string

  validation {
    condition     = length(var.domain_name) > 0
    error_message = "domain_name must be set — see the description for why this can't have a default."
  }
}

variable "route53_zone_id" {
  description = "Route 53 hosted zone ID for domain_name, used for ACM DNS validation and the ALB alias record. Leave empty if DNS for domain_name is hosted elsewhere — you'll need to add the ACM validation CNAME manually (see the acm_validation_records output)."
  type        = string
  default     = ""
}

variable "s3_bucket_name" {
  type    = string
  default = "email-agent-agentfiles"
}

variable "force_destroy" {
  description = "Allow `terraform destroy` to delete the agentfiles S3 bucket and ECR repo even if they still contain objects/images. Defaults to true (test/throwaway environments) — set to false for an environment you don't want torn down accidentally."
  type        = bool
  default     = true
}

variable "ssm_prefix" {
  description = "SSM Parameter Store path prefix under which per-user Gmail OAuth tokens are stored at runtime (e.g. /email-agent-token/<email>). Terraform only grants IAM access to this path — it doesn't manage the per-user parameters themselves, since they're created dynamically as users authenticate."
  type        = string
  default     = "/email-agent-token"
}

# --- Secrets, injected via SSM SecureString + the ECS "secrets" block.
# No defaults — pass at apply time (e.g. -var, or a gitignored .tfvars) so
# real values never land in version control or Terraform's own source. ---

variable "google_client_secret" {
  type      = string
  sensitive = true
}

variable "slack_webhook" {
  description = "Slack incoming webhook URL. Optional — the app treats an empty value as a no-op, so this can be left \"\" if you don't want Slack notifications yet."
  type        = string
  sensitive   = true
  default     = ""
}
