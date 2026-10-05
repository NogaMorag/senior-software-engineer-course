# Milestone M7: the Hookline chart installed by Terraform into its own namespace, next to the
# infrastructure that deploy/kind/infra.yaml created.

module "hookline" {
  source = "./modules/hookline"

  namespace       = var.namespace
  image_tag       = var.image_tag
  database_url    = var.database_url
  infra_namespace = var.infra_namespace
  worker_replicas = var.worker_replicas
}
