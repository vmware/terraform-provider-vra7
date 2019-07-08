provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

resource "vra7_deployment" "machine1" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"
}

resource "vra7_deployment" "machine2" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"
  depends_on        = [vra7_deployment.machine1]
}

