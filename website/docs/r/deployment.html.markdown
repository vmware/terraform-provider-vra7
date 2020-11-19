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

You can refer to the sample blueprint ([here](https://github.com/vmware/terraform-provider-vra7/tree/master/website/docs/r)) to understand how it is translated to following the terraform config

```hcl

resource "vra7_deployment" "this" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"
  description = "this description"
  reasons = "this reason"
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

To scale in and/or scale out the deployment created by the above configuration, change the cluster size within resource_configuration. For instance, scaling in Linux 1 and scaling out Linux 2

```hcl

resource_configuration  {
    component_name = "Linux 1"
    cluster = 1
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
    cluster = 3
    configuration = {
      cpu = 2
      memory = 1024
      storage = 8
    }
  }

```

To change lease, add expiry_date in your config. It is already present in the state file after the deployment is created. Refer state file for format. For instance, if the expiry_date after initial deployment in the state file was "2020-11-20T20:29:37.845Z", you can modify it as follows:

```hcl

resource "vra7_deployment" "this" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"
  description = "this description"
  reasons = "this reason"
  lease_days = 10
  expiry_date = "2020-11-25T20:29:37.845Z"
  deployment_configuration = {
    "blueprint_custom_property" = "This is a blueprint custom property"
  }

  // resource_configuration blocks as above examples
  resource_configuration  {}
  resource_configuration  {}
}

```

To reconfigure, change/add properties inside configuration block of resource_configuration block. For instance, in the initial deployment, to change cpu to 4 for both clusters, do the following:


```hcl

resource_configuration  {
    component_name = "Linux 1"
    cluster = 1
    configuration = {
      cpu = 4
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
    cluster = 3
    configuration = {
      cpu = 4
      memory = 1024
      storage = 8
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
* `lease_days` - (Optional) Number of lease days remaining for the deployment. NOTE: If this is not provided, the default lease_days in the catalog item will be configured. lease_days 0 means the lease never expires.
* `expiry_date` - (Optional) The date when the deployment will expire. To change lease, modify this field in main.tf. It has to be in the same format as in the state file. For e.g., "2020-11-25T20:29:37.845Z".
* `wait_timeout` - (Optional) Wait time out for the request. If the request is not completed within the timeout period, do a terraform refresh later to check the status of the request. 

## Attribute Reference

* `deployment_id` - The resource id of the deployment.
* `name` - The name of the deployment.
* `request_status` - The status of the catalog item request.
* `created_date` - The date when the deployment was created.
* `owners` - The owners of the deployment.

## Nested Blocks

### resource_configuration ###

This is a list of blocks that contains the machine resource level properties including the custom properties. Each resource_configuration block maps to a component in the blueprint/catalog_item. The sample blueprint has one vSphere machine resource/component called vSphereVM1. Properties of this machine can be specified in the config as shown in the example above. The properties like cpu, memory, storage, etc are generic machine properties and their is a custom property as well, called machine_property in the sample blueprint which is required at request time. The cluster property can be used to specify the number of machines corresponding that component. All the properties that are required during request, must be specified in the config file.

The following arguments for resource_configuration block are supported:

#### Argument Reference

* `component_name` - (Required) The name of the component/machine resource as in the blueprint/catalog_item
* `configuration` - (Optional) The machine resource level properties like cpu, memory, storage, custom properties, etc. can be added here. When fetching the state of the machine, this will be populated with a lot of information in the state file.
NOTE: To add an array property, refer to the security_tag value in example above.
* `cluster` - (Optional) Cluster size for this machine resource

#### Attribute Reference

* `instances` - List of the detailed state/view of the machine resources/instances/VMs within the deployment. This is a nested schema, discussed below
* `parent_resource_id` - ID of the deployment of which this machine is a part of
* `request_id` - ID of the catalog item request

### instances ###

* `resource_id` - ID of the machine resource
* `name` - Name of the resource
* `description` - Description of the resource
* `ip_address` - IP address of the machine
* `resource_type` - Type of resource. It can be a machine resource type (Infrastructure.Virtual) or a deployment type (composition.resource.type.deployment), etc.
* `properties` - Map of the instance/VM properties fetched from the deployment


### deployment_configuration ###

This block contains the deployment level properties including the custom properties and proprty groups. These are not a fixed set of properties but referred from the blueprint. From the example of the BasicSingleMachine blueprint, their is one custom property, called deployment_property which is required at request time. All the properties that are required during request, must be specified in the config file.
