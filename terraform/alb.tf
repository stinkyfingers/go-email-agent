resource "aws_security_group" "alb" {
  name        = "${var.app_name}-alb"
  description = "gmail_login ALB — public HTTP/HTTPS in, forwards to the ECS task"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb" "gmail_login" {
  name               = "${var.app_name}-gmail-login"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = data.aws_subnets.default.ids
}

resource "aws_lb_target_group" "gmail_login" {
  name        = "${var.app_name}-gmail-login"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "ip"

  health_check {
    path = "/login"
    # /login always 307-redirects to Google's consent URL — there's no
    # dedicated /health endpoint in gmail.go, so the health check has to
    # accept that redirect as "healthy" rather than a plain 200.
    matcher = "200,307"
  }
}

# ACM cert ARN is usable on the listener immediately (RequestCertificate
# returns it before validation completes) — so this listener works as soon
# as var.domain_name's DNS validation record is in place, with no further
# `apply` required once that happens.
resource "aws_acm_certificate" "gmail_login" {
  domain_name       = var.domain_name
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# Only wired up automatically when the domain's DNS is in Route 53 here. If
# not (route53_zone_id left ""), see the acm_validation_records output for
# the CNAME to add manually wherever that DNS actually lives.
resource "aws_route53_record" "gmail_login_cert_validation" {
  for_each = var.route53_zone_id != "" ? {
    for dvo in aws_acm_certificate.gmail_login.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  zone_id         = var.route53_zone_id
  name            = each.value.name
  type            = each.value.type
  records         = [each.value.record]
  ttl             = 60
  allow_overwrite = true
}

resource "aws_route53_record" "gmail_login" {
  count   = var.route53_zone_id != "" ? 1 : 0
  zone_id = var.route53_zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = aws_lb.gmail_login.dns_name
    zone_id                = aws_lb.gmail_login.zone_id
    evaluate_target_health = true
  }
}

resource "aws_lb_listener" "gmail_login_https" {
  load_balancer_arn = aws_lb.gmail_login.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = aws_acm_certificate.gmail_login.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.gmail_login.arn
  }
}

resource "aws_lb_listener" "gmail_login_http_redirect" {
  load_balancer_arn = aws_lb.gmail_login.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
