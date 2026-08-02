# terraform-provider-kea

Terraform / OpenTofu provider for the [kea-operator](https://git.unkin.net/unkin/kea-operator)
KeaAPI: it manages Kea DHCP subnets and PXE client classes as KeaSubnet /
KeaClientClass custom resources over the KeaAPI HTTP service.

## Provider configuration

```hcl
terraform {
  required_providers {
    kea = {
      source  = "git.unkin.net/unkin/kea"
      version = "0.0.1"
    }
  }
}

provider "kea" {
  endpoint = "https://keaapi.k8s.syd1.au.unkin.net"
  # token   = "..."   # defaults to $KEA_API_TOKEN
}
```

Both reads and writes require the bearer token. Set it out-of-band via the
`KEA_API_TOKEN` environment variable (managed in Vault) rather than in HCL.

## Resources

| Resource           | Purpose                                                        |
|--------------------|----------------------------------------------------------------|
| `kea_subnet`       | A DHCP subnet: pools, routers, DNS, PXE/boot options.          |
| `kea_clientclass`  | A PXE client class matching by test expr or architecture.      |

The `name` attribute is the stable id (the CR name) and forces replacement when
changed. `option_data` is a nested block of DHCP options (`name`, `code`,
`space`, and required `data`).

## Releases

Tagging `vX.Y.Z` builds `terraform-provider-kea_X.Y.Z_linux_amd64.zip` and
uploads it to the ArtifactAPI Terraform registry (`terraform-unkin` remote,
namespace `unkin/kea`), which serves it as a GPG-signed provider registry.

See `examples/` for full usage.
