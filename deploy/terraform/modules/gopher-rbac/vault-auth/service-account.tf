# Service Account
resource "kubernetes_service_account" "serviceaccount" {
  count = "${length(var.namespaces)}"

  metadata {
    labels      = "${var.labels}"
    annotations = "${var.annotations}"
    name        = "${element(var.namespaces, count.index)}-vault-auth"
    namespace   = "${element(var.namespaces, count.index)}"
  }
}

# ClusterRole Binding
resource "kubernetes_cluster_role_binding" "auth" {
  count = "${length(var.namespaces)}"

  metadata {
    name = "${element(var.namespaces, count.index)}-vault-auth"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "system:auth-delegator"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "${element(var.namespaces, count.index)}-vault-auth"
    namespace = "${element(var.namespaces, count.index)}"
  }
}
