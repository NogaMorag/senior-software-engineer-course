output "namespace" {
  value = kubernetes_namespace_v1.this.metadata[0].name
}

output "release" {
  value = {
    name     = helm_release.hookline.metadata.name
    chart    = "${helm_release.hookline.metadata.chart}-${helm_release.hookline.metadata.version}"
    revision = helm_release.hookline.metadata.revision
  }
}
