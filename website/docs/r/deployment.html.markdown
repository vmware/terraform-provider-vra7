---
layout: "vra7"
page_title: "VMware vRA7: vra7_deployment"
sidebar_current: "docs-vra7-resource-deployment"
description: |-
  Provides a VMware vRA7 deployment resource. This can be used to deploy vRA7 catalog items.
---

# vra7\_deployment

Provides a VMware vRA7 deployment resource. This can be used to deploy vRA7 catalog items.

## Example Usages

**Simple deployment of a vSphere machine with custom properties and a network profile:**

You can refer to the sample blueprint ([here](https://github.com/terraform-providers/terraform-provider-vra7/tree/master/website/docs/r)) to understand how it is translated to following the terraform config

```hcl
resource "vra7_deployment" "my_vra7_deployment" {
  count = 1
  catalog_item_name = "Basic Single Machine"
  description = "Test deployment"
  reasons = "Testing the vRA 7 Terraform plugin"

  deployment_configuration = {
    _leaseDays = "15"
    _number_of_instances = 2
    deployment_property = "custom deployment property"
  }
  resource_configuration = {
    vSphereVM1.cpu = 1
    vSphereVM1.memory = 2048
    vSphereVM1.machine_property = "machine custom property"
  }
  wait_timeout = 20
  businessgroup_name = "Development"
}
```

## Argument Reference

The following arguments are supported:

* `businessgroup_id` - (Optional) The id of the vRA business group to use for this deployment.
* `businessgroup_name` - (Optional) The name of the vRA business group to use for this deployment.
* `catalog_item_id` - (Optional) The id of the catalog item to deploy into vRA.
* `catalog_item_name` - (Optional) The name of the catalog item to deploy into vRA.
* `description` - (Optional) Description of the deployment
* `reasons` - (Optional) Reasons for requesting the deployment
* `deployment_configuration` - (Optional) The configuration of the deployment from the catalog item
* `deployment_configuration_list` - (Optional) The configuration of the deployment from the catalog item. The input values are comma separated values and will be passed as a list on the VRA api.
* `deployment_configuration_bool` - (Optional) The configuration of the deployment from the catalog item. The values should be booleans.
* `resource_configuration` - (Optional) The configuration of the individual components from the catalog item

## Nested Blocks

### deployment_configuration ###

This block contains the deployment level properties including the custom properties. These are not a fixed set of properties but referred from the blueprint. There are generic properties like _leaseDays, _number_of_instances, etc but they are optional and from the example of the BasicSingleMachine blueprint, their is one custom property, called deployment_property which is required at request time.
All the properties that are required during request, must be specified in the config file.

Variants exist: `deployment_configuration_list` and `deployment_configuration_bool`.

```
resource "vra7_deployment" "frontends" {
    catalog_item_name = "Create a Security Group"
    reasons = "Create frontends security groups"
    description = "Frontends 4"
    deployment_configuration_bool = {
        "commonName" = "dmz"
    }
    deployment_configuration_bool = {
        "isManaged" = true
        "isUniversal" = false
    }
    deployment_configuration_list = {
        "securityGroupIn" = "front1,front2"
    }
}

```

### resource_configuration ###

This block contains the machine resource level properties including the custom properties. These are not a fixed set of properties but referred from the blueprint. The sample blueprint has one vSphere machine resource called vSphereVM1. Properties of this machine can be specified in the config in the format "vSphereVM1.property_name". The properties like cpu, memory, storage, etc are generic machine properties and their is a custom property as well, called machine_property in the sample blueprint which is required at request time. There can be any number of machines and same format has to be followed to specify properties of other machines as well.
All the properties that are required during request, must be specified in the config file.


### More examples ###

**Simple deployment of a CentOS Linux host with 2 CPU's:**

```hcl
resource "vra7_deployment" "machine" {
  count            = 1
  catalog_item_name = "CentOS 7.0 x64"
  resource_configuration = {
    Linux.cpu = "2"
  }
}
```

**Catalog "multi_machine_catalog" contains Linux, Windows and http (apache) designs:**

```hcl
resource "vra7_deployment" "resource_1" {
  count            = 1
  catalog_item_name = "multi_machine_catalog"
  resource_configuration = {
    Windows.cpu = "2"                //Windows Machine CPU
    Linux.cpu = "2"                  //Linux Machine CPU
    http.hostname = "xyz.com"        //HTTP (apache) hostname
    http.network_mode = "bridge"     //HTTP (apache) network mode
  }
}
```

**Showing the use of `depends_on` between two deployments:**

Here the second resource, machine2 is dependent on the resource, machine1. So, machine2 will be provisioned after machine1.

```hcl
resource "vra7_deployment" "machine1" {
  count            = 1
  catalog_item_name = "CentOS 7.0 x64"
}

resource "vra7_deployment" "machine2" {
  count            = 1
  catalog_item_name = "CentOS 7.0 x64"
  depends_on = ["vra7_deployment.machine1"]
}
```
