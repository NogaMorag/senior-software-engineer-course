variable "namespace" {
  description = "Namespace that the module creates and installs into."
  type        = string
}

variable "release_name" {
  description = "Helm release name."
  type        = string
  default     = "shorty"
}

variable "replicas" {
  description = "Replica count (the chart's replicaCount)."
  type        = number
  default     = 2

  validation {
    condition     = var.replicas >= 1 && var.replicas <= 10
    error_message = "The replicas value must be between 1 and 10, as in the chart's values.schema.json."
  }
}

variable "log_level" {
  description = "The chart's config.logLevel."
  type        = string
  default     = "info"
}

variable "image_tag" {
  description = "The chart's image.tag."
  type        = string
}
