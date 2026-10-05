# The shorty module: a namespace that Terraform owns, and the chapter 21 chart
# installed into it from the local copy in ./chart.

# TODO 5: a kubernetes_namespace_v1 "this" named var.namespace.

# TODO 6: a helm_release "shorty": name var.release_name, namespace from TODO 5,
# chart from the local path "${path.module}/chart", values as one yamlencode()
# document with replicaCount, image.tag and config.logLevel, wait = true and a
# timeout of 120 seconds.
