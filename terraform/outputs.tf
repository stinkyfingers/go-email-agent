output "ecr_repository_url" {
  value = aws_ecr_repository.app.repository_url
}

output "gmail_login_alb_dns_name" {
  value = aws_lb.gmail_login.dns_name
}

output "gmail_login_url" {
  value = "https://${var.domain_name}/login"
}

output "google_oauth_redirect_url" {
  description = "Add this exact URL to the OAuth client's Authorized redirect URIs in Google Cloud Console — Terraform can't do this step."
  value       = "https://${var.domain_name}/callback"
}

output "acm_validation_records" {
  description = "If route53_zone_id is unset, add these CNAME records wherever domain_name's DNS is actually hosted to complete ACM validation."
  value = [
    for dvo in aws_acm_certificate.gmail_login.domain_validation_options : {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  ]
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.app.name
}
