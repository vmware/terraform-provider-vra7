provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

resource "vra7_deployment" "this" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"
  description = "this description"
  reason = "this reason"
  lease_days = 10

  deployment_configuration = {
    "blueprint_custom_property" = "This is a blueprint custom property"
    "businessGroups" = <<EOF
        [
          "bgTest1",
          "bgTest2"
        ]
        EOF
  }
  
  resource_configuration  {
    component_name = "Linux 1"
    cluster = 2
    configuration = {
      cpu = 2
      memory = 2048
      custom_property = "VM custom property"
      security_tag = <<EOF
        [
          "dev_sg",
          "prod_sg"
        ]
        EOF
    }
  }

  resource_configuration  {
    component_name = "Linux 2"
    configuration = {
      cpu = 2
      memory = 1024
      storage = 8
    }
  }
}

