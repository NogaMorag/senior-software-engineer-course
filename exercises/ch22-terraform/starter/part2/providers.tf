provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

# TODO 4: a helm provider configured for the same kubeconfig and context.
# In helm provider v3, kubernetes is an attribute that takes an object.
