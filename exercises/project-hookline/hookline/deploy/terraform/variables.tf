variable "kubeconfig" {
  description = "Path to the kubeconfig file of the kind cluster (scripts/kind-smoke.sh writes /tmp/hookline-kind.kubeconfig)."
  type        = string
  default     = "/tmp/hookline-kind.kubeconfig"
}

variable "kube_context" {
  description = "The kubeconfig context of the kind cluster."
  type        = string
  default     = "kind-hookline"
}

variable "namespace" {
  description = "Namespace that Terraform creates for this Hookline release."
  type        = string
  default     = "hookline-tf"
}

variable "image_tag" {
  description = "Tag of the hookline image already loaded into kind."
  type        = string
  default     = "dev"
}

variable "database_url" {
  description = "DATABASE_URL for the release; it becomes a Secret. Set it in terraform.tfvars, never in Git."
  type        = string
  sensitive   = true
}

variable "infra_namespace" {
  description = "Namespace of PostgreSQL, Redis and Redpanda (deploy/kind/infra.yaml)."
  type        = string
  default     = "hookline"
}

variable "worker_replicas" {
  description = "Worker replicas; the one value this exercise changes between applies."
  type        = number
  default     = 1
}
