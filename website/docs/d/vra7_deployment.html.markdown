---
layout: "vra7"
page_title: "VMware vRA7: vra7_deployment"
sidebar_current: "docs-vra7-resource-deployment"
description: |-
  Provides a VMware vRA7 deployment data source. This can be used to get a vra7_deployment
---

# Data Source vra7\_deployment

Provides a VMware vRA7 deployment data source. This can be used to get a vra7_deployment

## Example Usages

### Filter by deployment id

```hcl
data "vra7_deployment" "this" {
  deployment_id = "a0967161-d80f-220c-9c7a-5892025bc3ce"
}
```
### Filter by catalog item request id

```hcl
data "vra7_deployment" "this" {
  id = "a0967161-d80f-220c-9c7a-5892025bc3ce"
}
```

## Argument Reference

The following arguments are supported:
* `id` - The catalog item request id.
* `deployment_id` - The resource id of the deployment. 

## Attribute Reference

* `businessgroup_id` - The id of the vRA business group to use for this deployment.
* `businessgroup_name` - The name of the vRA business group to use for this deployment.
* `catalog_item_id` - The id of the catalog item to deploy into vRA.
* `catalog_item_name` - The name of the catalog item to deploy into vRA.
* `description` - Description of the deployment.
* `reasons` - Reasons for requesting the deployment.
* `deployment_configuration` - The configuration of the deployment from the catalog item. All blueprint custom properties including property groups can be added to this block. This property is discussed in detail below.
* `resource_configuration` - The configuration of the individual components from the catalog item. This property is discussed in detail below.
* `lease_days` - Number of lease days remaining for the deployment. NOTE: lease_days 0 means the lease never expires.
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

#### Attribute Reference

* `component_name` - The name of the component/machine resource as in the blueprint/catalog_item
* `resource_state` - The detailed state/view of the machine resources within the deployment
* `cluster` - Cluster size for this machine resource
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
