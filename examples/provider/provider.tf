terraform {
  required_providers {
    kea = {
      source = "IUCLabs/kea"
    }
  }
}

# The token defaults to the KEA_API_TOKEN environment variable.
provider "kea" {
  endpoint = "https://keaapi.example.com"
}
