output "namespace" {
  value = module.hookline.namespace
}

output "release" {
  description = "Release name, chart version and revision, as Helm records them."
  value       = module.hookline.release
}

output "port_forward" {
  description = "The command that reaches the API from your machine."
  value       = "kubectl -n ${module.hookline.namespace} port-forward svc/hookline-api 8080:80"
}
