# Providers are configured only here, in the root module (chapter 22); the module only declares them.
provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

# Helm provider v3: "kubernetes" is an attribute that takes an object (= { }), not a nested block.
provider "helm" {
  kubernetes = {
    config_path    = var.kubeconfig
    config_context = var.kube_context
  }
}
