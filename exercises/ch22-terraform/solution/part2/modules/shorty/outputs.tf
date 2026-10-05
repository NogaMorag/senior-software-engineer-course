output "namespace" {
  value = kubernetes_namespace_v1.this.metadata[0].name
}

output "service_name" {
  description = "The chart names the Service after the release (see _helpers.tpl)."
  value       = helm_release.shorty.name
}

output "release" {
  value = {
    name     = helm_release.shorty.metadata.name
    chart    = "${helm_release.shorty.metadata.chart}-${helm_release.shorty.metadata.version}"
    revision = helm_release.shorty.metadata.revision
  }
}
