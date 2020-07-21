resource "kubernetes_cluster_role" "clusterrole" {
  metadata {
    name = "${var.cluster_role_name}"
  }

  rule {
    api_groups = [""]
    resources  = ["secrets"]
    verbs      = ["get", "create", "update"]
  }
}
