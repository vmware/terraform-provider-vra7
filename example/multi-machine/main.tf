provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

# Catalog "multi_machine_catalog" contains Linux, Windows and http (apache) designs.
resource "vra7_deployment" "resource_1" {
  count             = 1
  catalog_item_name = "multi_machine_catalog"
  resource_configuration = {
    "Windows.cpu"       = "2"       //Windows Machine CPU
    "Linux.cpu"         = "2"       //Linux Machine CPU
    "http.hostname"     = "xyz.com" //HTTP (apache) hostname
    "http.network_mode" = "bridge"  //HTTP (apache) network mode
  }
}

