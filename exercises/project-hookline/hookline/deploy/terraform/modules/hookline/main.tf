# The Hookline module: a namespace Terraform owns, the database Secret, and the chart from the repository
# (a local path, the same chart helm install uses; not a copy).

resource "kubernetes_namespace_v1" "this" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

resource "kubernetes_secret_v1" "database" {
  metadata {
    name      = "hookline-database"
    namespace = kubernetes_namespace_v1.this.metadata[0].name
  }
  data = {
    url = var.database_url
  }
}

locals {
  infra = "${var.infra_namespace}.svc.cluster.local"
}

resource "helm_release" "hookline" {
  name      = var.release_name
  namespace = kubernetes_namespace_v1.this.metadata[0].name
  chart     = "${path.module}/../../../helm/hookline"

  # Values as one YAML document, exactly like a -f file for helm install.
  values = [yamlencode({
    image        = { tag = var.image_tag }
    devMode      = true
    redisAddr    = "redis.${local.infra}:6379"
    kafkaBrokers = "redpanda.${local.infra}:9092"
    api          = { replicas = 1 }
    relay        = { replicas = 1 }
    dispatcher   = { replicas = 1 }
    worker       = { replicas = var.worker_replicas }
  })]

  wait    = true # apply returns only when the pods are ready
  timeout = 300  # seconds

  depends_on = [kubernetes_secret_v1.database] # the pre-install hook Job reads it
}
