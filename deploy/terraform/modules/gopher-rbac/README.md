# Sample configuration of the gopher app RBAC modules

### This will be a shared clusterrole and need not to be called always.
```hcl
module "clusterrole" {
  source            = "../../../modules/gopher-rbac/clusterrole"
  cluster_role_name = "vault-gopher"
}
```

### Namespaces variable is type of list, we can create all binding and service account of all namespace in envs (sit, jackal, lion and eagle).
```hcl
module "sre" {
  source            = "../../../modules/gopher-rbac/gopher-auth"
  namespaces        = ["sit-sre"]
  cluster_role_name = "vault-gopher"
}
```

# Sample configuration of the gopher auth RBAC module

### Call the module on each environment (sit, jackal, lion and eagle), namespaces variable is type of list.
- For namespace "sit-sre"
```hcl
module "sit-sre-auth" {
  source       = "../../../modules/gopher-rbac/vault-auth"
  namespaces   = ["sit-sre"]
}
```