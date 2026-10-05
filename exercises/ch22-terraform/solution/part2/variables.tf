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
  description = "Namespace for the shorty release."
  type        = string
  default     = "shorty-tf"
}

variable "replicas" {
  description = "Replica count passed to the chart."
  type        = number
  default     = 2
}

variable "log_level" {
  description = "SHORTY_LOG_LEVEL, passed to the chart as config.logLevel."
  type        = string
  default     = "info"
}

variable "image_tag" {
  description = "Tag of the shorty image already loaded into kind."
  type        = string
  default     = "ch21"
}
