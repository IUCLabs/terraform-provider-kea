# Terraform Provider Kea

This provider manages [ISC Kea](https://www.isc.org/kea/) DHCP configuration
with Terraform or OpenTofu. It talks to a KeaAPI HTTP service and manages DHCP
subnets and PXE client classes.

See the [provider documentation](docs/index.md) for all resources and arguments.

## Quick Start

Prebuilt provider archives are on the [releases page](https://github.com/IUCLabs/terraform-provider-kea/releases).

```hcl
terraform {
  required_providers {
    kea = {
      source = "IUCLabs/kea"
    }
  }
}

provider "kea" {
  endpoint = "https://keaapi.example.com"
  # token defaults to the KEA_API_TOKEN environment variable
}

resource "kea_subnet" "lab" {
  name    = "lab"
  subnet  = "192.0.2.0/24"
  pools   = ["192.0.2.100-192.0.2.200"]
  routers = ["192.0.2.1"]
}
```

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) 1.x or [OpenTofu](https://opentofu.org/)
- [Go](https://go.dev/doc/install) 1.25+ (to build the provider)

## Development

```sh
make build   # build the provider binary
make test    # run unit tests
make docs    # regenerate docs/ from the schema and examples/
```

## License

[MIT](LICENSE)
