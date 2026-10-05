# OPTIONAL, section 22.7. The author validated this configuration but never
# applied it: it needs your own AWS account, and an account can cost money.
terraform {
  required_version = ">= 1.16.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.67.0"
    }
  }
}
