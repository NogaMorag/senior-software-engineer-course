# Day 65: one module call creates a namespace and the shorty release in it.

module "shorty" {
  source = "./modules/shorty"

  namespace = var.namespace
  replicas  = var.replicas
  log_level = var.log_level
  image_tag = var.image_tag
}
