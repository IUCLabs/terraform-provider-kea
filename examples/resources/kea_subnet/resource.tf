resource "kea_subnet" "lab" {
  name        = "lab"
  cluster_ref = "primary"
  subnet      = "192.0.2.0/24"

  pools          = ["192.0.2.100-192.0.2.200"]
  routers        = ["192.0.2.1"]
  dns_servers    = ["192.0.2.53"]
  domain_name    = "lab.example.com"
  valid_lifetime = 3600

  # PXE boot options
  next_server    = "192.0.2.1"
  boot_file_name = "pxelinux.0"
  client_classes = ["pxeclients"]

  option_data = [
    {
      name = "tftp-server-name"
      code = 66
      data = "192.0.2.1"
    },
  ]
}
