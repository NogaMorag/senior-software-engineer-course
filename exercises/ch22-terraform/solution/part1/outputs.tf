output "namespace" {
  description = "The namespace Terraform created."
  value       = kubernetes_namespace_v1.demo.metadata[0].name
}

output "config_map" {
  description = "Namespace and name of the ConfigMap, as kubectl expects them."
  value       = "${kubernetes_config_map_v1.settings.metadata[0].namespace}/${kubernetes_config_map_v1.settings.metadata[0].name}"
}
