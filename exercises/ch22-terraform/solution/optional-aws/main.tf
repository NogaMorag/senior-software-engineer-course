provider "aws" {
  region = var.region

  default_tags {
    tags = {
      course     = "senior-course"
      chapter    = "22"
      managed-by = "terraform"
    }
  }
}

# One private, versioned, encrypted bucket: the kind of bucket a team keeps
# Terraform state in (see backend-s3.tf.example).
resource "aws_s3_bucket" "state" {
  bucket        = var.bucket_name
  force_destroy = true # lets terraform destroy remove a bucket that still has objects; fine for a course bucket only
}

resource "aws_s3_bucket_versioning" "state" {
  bucket = aws_s3_bucket.state.id
  versioning_configuration {
    status = "Enabled" # an overwritten state file can be recovered
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "state" {
  bucket                  = aws_s3_bucket.state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
