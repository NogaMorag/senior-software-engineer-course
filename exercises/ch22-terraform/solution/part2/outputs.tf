output "namespace" {
  value = module.shorty.namespace
}

output "release" {
  description = "Release name, chart version and revision, as Helm records them."
  value       = module.shorty.release
}

output "port_forward" {
  description = "The command that reaches the service from your machine."
  value       = "kubectl -n ${module.shorty.namespace} port-forward svc/${module.shorty.service_name} 8080:80"
}
