variable "namespaces" {
  type = "list"
}

variable "labels" {
  type    = "map"
  default = {}
}

variable "annotations" {
  type    = "map"
  default = {}
}
