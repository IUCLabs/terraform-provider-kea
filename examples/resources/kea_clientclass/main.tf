resource "kea_clientclass" "pxeclients" {
  name = "pxeclients"

  # Match legacy BIOS PXE clients by vendor-class-identifier.
  test = "substring(option[60].hex,0,9) == 'PXEClient'"

  next_server    = "10.0.0.1"
  boot_file_name = "ipxe.efi"
}

resource "kea_clientclass" "efi_x64" {
  name = "efi-x64"

  # Match by client architecture type instead of a test expression.
  arch_hex       = ["0007", "0009"]
  boot_file_name = "ipxe.efi"
}
