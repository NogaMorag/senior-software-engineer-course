# The shorty module: a namespace that Terraform owns, and the chapter 21 chart
# installed into it from the local copy in ./chart.

resource "kubernetes_namespace_v1" "this" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

resource "helm_release" "shorty" {
  name      = var.release_name
  namespace = kubernetes_namespace_v1.this.metadata[0].name
  chart     = "${path.module}/chart" # a local path, not a repository

  # Values as one YAML document, exactly like a -f file for helm install.
  values = [yamlencode({
    replicaCount = var.replicas
    image        = { tag = var.image_tag }
    config       = { logLevel = var.log_level }
  })]

  wait    = true # apply returns only when the pods are ready
  timeout = 120  # seconds
}
