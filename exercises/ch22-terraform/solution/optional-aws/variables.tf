variable "region" {
  description = "AWS region for the bucket."
  type        = string
  default     = "eu-central-1"
}

variable "bucket_name" {
  description = "Globally unique bucket name, for example senior-course-tfstate-<your initials>-<digits>."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", var.bucket_name))
    error_message = "Bucket names are 3 to 63 characters: lowercase letters, digits, dots and hyphens."
  }
}
