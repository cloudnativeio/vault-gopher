# Service Account
resource "kubernetes_service_account" "serviceaccount" {
  count = "${length(var.namespaces)}"

  metadata {
    labels      = "${var.labels}"
    annotations = "${var.annotations}"
    name        = "${element(var.namespaces, count.index)}-${var.cluster_role_name}"
    namespace   = "${element(var.namespaces, count.index)}"
  }

  automount_service_account_token = true
}

# Role Binding
resource "kubernetes_role_binding" "rolebinding" {
  count = "${length(var.namespaces)}"

  metadata {
    name      = "${element(var.namespaces, count.index)}-${var.cluster_role_name}"
    namespace = "${element(var.namespaces, count.index)}"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "${var.cluster_role_name}"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "${element(var.namespaces, count.index)}-${var.cluster_role_name}"
    namespace = "${element(var.namespaces, count.index)}"
  }
}
