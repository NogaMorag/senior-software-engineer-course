variable "namespace" {
  description = "Namespace that the module creates and installs into."
  type        = string
}

variable "release_name" {
  description = "Helm release name; the chart names its objects after it."
  type        = string
  default     = "hookline"
}

variable "image_tag" {
  description = "The chart's image.tag."
  type        = string
}

variable "database_url" {
  description = "DATABASE_URL, stored in the Secret the chart reads."
  type        = string
  sensitive   = true
}

variable "infra_namespace" {
  description = "Namespace of the shared infrastructure, for the service addresses."
  type        = string
}

variable "worker_replicas" {
  description = "The chart's worker.replicas."
  type        = number

  validation {
    condition     = var.worker_replicas >= 0 && var.worker_replicas <= 20
    error_message = "The worker_replicas value must be between 0 and 20, as in the chart's values.schema.json."
  }
}
