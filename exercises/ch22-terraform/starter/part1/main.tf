# Day 64: the kubernetes provider against the kind cluster "ch22".

provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

# TODO 1: a kubernetes_namespace_v1 resource named "demo" whose metadata name is
# var.namespace, labelled app.kubernetes.io/managed-by = terraform.

# TODO 2: a kubernetes_config_map_v1 resource named "settings": name
# shorty-settings, in the namespace of TODO 1 (use a reference to the namespace
# resource, not the variable, so Terraform orders the two), with the data keys
# SHORTY_BASE_URL = http://localhost:8080 and SHORTY_LOG_LEVEL = var.log_level.
