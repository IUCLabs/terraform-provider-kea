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
  # token defaults to the KEA_API_TOKEN environment variable
}

resource "kea_clientclass" "pxeclients" {
  name = "pxeclients"
  test = "substring(option[60].hex,0,9) == 'PXEClient'"

  next_server    = "10.0.0.1"
  boot_file_name = "ipxe.efi"
}

resource "kea_subnet" "lab" {
  name        = "lab"
  cluster_ref = "primary"
  subnet      = "10.0.0.0/24"

  pools          = ["10.0.0.100-10.0.0.200"]
  routers        = ["10.0.0.1"]
  dns_servers    = ["10.0.0.53"]
  domain_name    = "lab.unkin.net"
  client_classes = [kea_clientclass.pxeclients.name]
  valid_lifetime = 3600

  option_data = [
    {
      name = "tftp-server-name"
      code = 66
      data = "10.0.0.1"
    },
  ]
}
