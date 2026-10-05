provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

# Helm provider v3: "kubernetes" is an attribute that takes an object (= { }),
# not a nested block as in provider v2.
provider "helm" {
  kubernetes = {
    config_path    = var.kubeconfig
    config_context = var.kube_context
  }
}
