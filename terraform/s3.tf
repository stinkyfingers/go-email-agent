resource "aws_s3_bucket" "agentfiles" {
  bucket        = var.s3_bucket_name
  force_destroy = var.force_destroy
}

resource "aws_s3_bucket_versioning" "agentfiles" {
  bucket = aws_s3_bucket.agentfiles.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "agentfiles" {
  bucket = aws_s3_bucket.agentfiles.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "agentfiles" {
  bucket                  = aws_s3_bucket.agentfiles.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
