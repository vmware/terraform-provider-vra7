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
resource "vra7_deployment" "this" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"
  description = "this description"
  reason = "this reason"
  lease_days = 10

  deployment_configuration = {
    "blueprint_custom_property" = "This is a blueprint custom property"
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
```

## Argument Reference

The following arguments are supported:

* `businessgroup_id` - (Optional) The id of the vRA business group to use for this deployment. Either businessgroup_id or businessgroup_name is required.
* `businessgroup_name` - (Optional) The name of the vRA business group to use for this deployment. Either businessgroup_id or businessgroup_name is required.
* `catalog_item_id` - (Optional) The id of the catalog item to deploy into vRA. Either catalog_item_id or catalog_item_name is required.
* `catalog_item_name` - (Optional) The name of the catalog item to deploy into vRA. Either catalog_item_id or catalog_item_name is required.
* `description` - (Optional) Description of the deployment.
* `reasons` - (Optional) Reasons for requesting the deployment.
* `deployment_configuration` - (Optional) The configuration of the deployment from the catalog item. All blueprint custom properties including property groups can be added to this block. This property is discussed in detail below.
* `resource_configuration` - (Optional) The configuration of the individual components from the catalog item. This property is discussed in detail below.
* `lease_days` - (Optional) Number of lease days remaining for the deployment. NOTE: lease_days 0 means the lease never expires.
* `wait_timeout` - (Optional) Wait time out for the request. If the request is not completed within the timeout period, do a terraform refresh later to check the status of the request. 

## Attribute Reference

* `deployment_id` - The resource id of the deployment.
* `name` - The name of the deployment.
* `lease_start` - Start date of the lease.
* `lease_end` - End date of the lease.
* `request_status` - The status of the catalog item request.
* `date_created` - The date when the deployment was created.
* `last_updated` - The date when the deployment was last updated after Day 2 operations.
* `tenant_id` - The id of the tenant.
* `owners` - The owners of the deployment.

## Nested Blocks

### resource_configuration ###

This is a list of blocks that contains the machine resource level properties including the custom properties. Each resource_configuration block corresponds to a component in the blueprint/catalog_item. The sample blueprint has one vSphere machine resource/component called vSphereVM1. Properties of this machine can be specified in the config as shown in the example above. The properties like cpu, memory, storage, etc are generic machine properties and their is a custom property as well, called machine_property in the sample blueprint which is required at request time. There can be any number of machines and same format has to be followed to specify properties of other machines as well.All the properties that are required during request, must be specified in the config file.

The following arguments for resource_configuration block are supported:

#### Argument Reference

* `component_name` - (Required) The name of the component/machine resource as in the blueprint/catalog_item
* `configuration` - (Optional) The machine resource level properties like cpu, memory, storage, custom properties, etc. can be added here. When fetching the state of the machine, this will be populated with a lot of information in the state file.
NOTE: To add an array property, refer to the security_tag value in example above.
* `cluster` - (Optional) Cluster size for this machine resource

#### Attribute Reference

* `resource_id` - ID of the machine resource
* `name` - Name of the resource
* `description` - Description of the resource
* `parent_resource_id` - ID of the deployment of which this machine is a part of
* `ip_address` - IP address of the machine
* `request_id` - ID of the catalog item request
* `request_state` - Current state of the request. It can be FAILED, IN_PROGRESS, SUCCESSFUL, etc.
* `resource_type` - Type of resource. It can be a machine resource type (Infrastructure.Virtual) or a deployment type (composition.resource.type.deployment), etc.
* `status` - Status of the machine. It can be On, Off, Unprovisioned, etc.
* `date_created` - The date when the resource was created.
* `last_updated` - The date when the resource was last updated after Day 2 operations. 


### deployment_configuration ###

This block contains the deployment level properties including the custom properties and proprty groups. These are not a fixed set of properties but referred from the blueprint. From the example of the BasicSingleMachine blueprint, their is one custom property, called deployment_property which is required at request time. All the properties that are required during request, must be specified in the config file.
