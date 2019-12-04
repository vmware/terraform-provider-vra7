provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

resource "vra7_deployment" "this" {
  catalog_item_name = var.catalog_item_name
  reasons           = var.description
  description       = var.description

  count = var.count

  deployment_configuration = {
    "VirtualMachine.Disk1.Size" = var.extra_disk
  }

  resource_configuration = {
    component_name = "Machine"
    configuration = {
      description = var.description
      cpu         = var.cpu
      memory      = var.ram
    }
  }

  wait_timeout = var.wait_timeout

  // Connection settings
  // Connection settings
  connection {
    host     = self.resource_configuration[*].ip_address
    user     = var.ssh_user
    password = var.ssh_password
    type     = ssh
  }

  // Extend volume to second disk
  // Extend volume to second disk
  provisioner "remote-exec" {
    inline = [
      "pvcreate /dev/sdb",
      "vgextend VolGroup00 /dev/sdb",
      "lvextend -l +100%FREE /dev/mapper/VolGroup00-rootLV",
      "resize2fs /dev/mapper/VolGroup00-rootLV",
    ]
  }
}

