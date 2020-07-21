variable "namespaces" {
  type = "list"
}

variable "cluster_role_name" {}

variable "labels" {
  type    = "map"
  default = {}
}

variable "annotations" {
  type    = "map"
  default = {}
}
