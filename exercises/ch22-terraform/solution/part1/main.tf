# Day 64: the kubernetes provider against the kind cluster "ch22".

provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

resource "kubernetes_namespace_v1" "demo" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

resource "kubernetes_config_map_v1" "settings" {
  metadata {
    name      = "shorty-settings"
    namespace = kubernetes_namespace_v1.demo.metadata[0].name # a reference: Terraform creates the namespace first
  }

  data = {
    SHORTY_BASE_URL  = "http://localhost:8080"
    SHORTY_LOG_LEVEL = var.log_level
  }
}
