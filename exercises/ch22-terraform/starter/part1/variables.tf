variable "kubeconfig" {
  description = "Path to the kubeconfig file that kind writes."
  type        = string
  default     = "~/.kube/config"
}

variable "kube_context" {
  description = "The kubeconfig context of the kind cluster."
  type        = string
  default     = "kind-ch22"
}

variable "namespace" {
  description = "Namespace that Terraform creates and owns."
  type        = string
  default     = "tf-demo"
}

variable "log_level" {
  description = "Value of SHORTY_LOG_LEVEL in the ConfigMap."
  type        = string
  default     = "info"

  validation {
    condition     = contains(["debug", "info", "warn", "error"], var.log_level)
    error_message = "The log_level must be one of debug, info, warn, error."
  }
}
