resource "kea_subnet" "lab" {
  name        = "lab"
  cluster_ref = "primary"
  subnet      = "10.0.0.0/24"

  pools          = ["10.0.0.100-10.0.0.200"]
  routers        = ["10.0.0.1"]
  dns_servers    = ["10.0.0.53"]
  domain_name    = "lab.unkin.net"
  valid_lifetime = 3600

  # PXE boot options
  next_server    = "10.0.0.1"
  boot_file_name = "pxelinux.0"
  client_classes = ["pxeclients"]

  option_data = [
    {
      name = "tftp-server-name"
      code = 66
      data = "10.0.0.1"
    },
  ]
}
