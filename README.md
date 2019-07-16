Terraform provider for VMware vRealize Automation 7
==================

- Website: https://www.terraform.io
- Documentation: https://www.terraform.io/docs/providers/vra7/index.html
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Introduction
------------

A self-contained deployable integration between Terraform and vRealize Automation (vRA) which allows Terraform users to request/provision entitled vRA catalog items using Terraform. Supports Terraform destroying vRA provisioned resources.

Requirements
------------

* [Terraform 0.9 or above](https://www.terraform.io/downloads.html)
* [Go Language 1.11.4 or above](https://golang.org/dl/)
* [vRealize Automation 7.4 or above](https://www.vmware.com/products/vrealize-automation.html)

Using the provider
----------------------

See the [vra7 documentation](https://www.terraform.io/docs/providers/vra7/index.html) to get started using the vRealize Automation 7 provider.

Upgrading the provider
----------------------

The vra7 provider doesn't upgrade automatically once you've started using it. After a new release you can run 

```bash
terraform init -upgrade
```

to upgrade to the latest stable version of the vra7 provider. See the [Terraform website](https://www.terraform.io/docs/configuration/providers.html#provider-versions)
for more information on provider upgrades, and how to set version constraints on your provider.

## Configure
The VMware vRA terraform configuration file contains two objects.

### Provider
This part contains service provider details.

Provider block contains four mandatory fields:
* `username` - vRA portal username
* `password` - vRA portal password
* `tenant` - vRA portal tenant
* `host` - End point of REST API
* `insecure` - In case of self-signed certificates, default value is false

**Example:**
```
    provider "vra7" {
      username = "vRAUser1@vsphere.local"
      password = "password123!"
      tenant = "corp.local.tenant"
      host = "http://myvra.example.com/"
      insecure = false
    }

```

### Resource
This part contains any resource that can be deployed on that service provider.
For example, in our case machine blueprint, software blueprint, complex blueprint, network, etc.

**Syntax:**
```
resource "vra7_deployment" "<resource_name1>" {
}
```

The resource block contains mandatory and optional fields.

**Mandatory:**

One of catalog_item_name or catalog_item_id must be specified in the resource configuration.
* `catalog_item_name` - catalog_item_name is a field which contains valid catalog item name from your vRA
* `catalog_item_id` - catalog_item_id is a field which contains a valid catalog item id from your vRA

**Optional:**
* `description` - This is an optional field. You can specify a description for your deployment.
* `reasons` - This is an optional field. You can specify the reasons for this deployment.
* `businessgroup_id` - This is an optional field. You can specify a different Business Group ID from what provided by default in the template request, provided that your account is allowed to do it.
* `businessgroup_name` - This is an optional field. You can specify a different Business Group name from what provided by default in the template request, provided that your account is allowed to do it.
* `count` - This field is used to create replicas of resources. If count is not provided then it will be considered as 1 by default.
* `deployment_configuration` - This is an optional field. It can be used to specify deployment level properties like _leaseDays, _number_of_instances or any custom properties of the deployment. Key is any field name of catalog item and value is any valid user input to the respective field..
* `resource_configuration` - This is an optional field. If blueprint properties have default values or no mandatory property value is required then you can skip this field from terraform configuration file. This field contains user inputs to catalog services. Value of this field is in key value pair. Key is service.field_name and value is any valid user input to the respective field.
* `wait_timeout` - This is an optional field with a default value of 15. It defines the time to wait (in minutes) for a resource operation to complete successfully.


**Example 1:**
```
resource "vra7_deployment" "example_machine1" {
  catalog_item_name = "CentOS 6.3"
  reasons      = "I have some"
  description  = "deployment via terraform"
   resource_configuration = {
         "Linux.cpu" = "1"
         "Windows2008R2SP1.cpu" =  "2"
         "Windows2012.cpu" =  "4"
         "Windows2016.cpu" =  "2"
     }
     deployment_configuration = {
         "_leaseDays" = "5"
     }
     count = 3
}
```

**Example 2:**
```
resource "vra7_deployment" "example_machine2" {
  catalog_item_id = "e5dd4fba7f96239286be45ed"
   resource_configuration = {
         "Linux.cpu" = "1"
         "Windows2008.cpu" =  "2"
         "Windows2012.cpu" =  "4"
         "Windows2016.cpu" =  "2"
     }
     count = 4
}

```

Save this configuration in `main.tf` in a path where the binary is placed.

## Execution
These are the Terraform commands that can be used for the vRA plugin:
* `terraform init` - The init command is used to initialize a working directory containing Terraform configuration files.
* `terraform plan` - Plan command shows plan for resources like how many resources will be provisioned and how many will be destroyed.
* `terraform apply` - apply is responsible to execute actual calls to provision resources.
* `terraform refresh` - By using the refresh command you can check the status of the request.
* `terraform show` - show will set a console output for resource configuration and request status.
* `terraform destroy` - destroy command will destroy all the  resources present in terraform configuration file.

Navigate to the location where `main.tf` and binary are placed and use the above commands as needed.

Building the provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-vra7`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-vra7
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-vra7
$ make build
```

Developing the provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11.4+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-vra7
...
```

# Scripts

For some older installations prior to v0.2.0 the **update_resource_state.sh** may need to be run.

There are few changes in the way the terraform config file is written.
1. The resource name is renamed to vra7_deployment from vra7_resource.
2. catalog_name is renamed to catalog_item_name and catalog_id is renamed to catalog_item_id.
3. General properties of deployment like description and reasons are to be specified at the resource level map instead of deployment_configuration.
4. catalog_configuration map is removed.
5. Custom/optional properties of deployment are to be specified in deployment_configuration instead of catalog_configuration.

These changes in the config file will lead to inconsistency in the `terraform.tfstate` file of the existing resources provisioned using terraform.
The existing state files can be converted to the new format using the script, `update_resource_state.sh` under the scripts folder.

Note: This script will only convert the state file. The changes to the config file(.tf file) still needs to be done manually.

## How to use the script

1. Copy the script, `script/update_resource_state.sh` in the same directory as your terraform.tfstate file.
2. Change the permission of the script, for example `chmod 0700 update_resource_state.sh`.
3. Run the script, `./update_resource_state.sh`.
4. The terraform.tfstate will be updated to the new format and a back-up of the old file is saved as terraform.tfstate_back

Contributing
------------

Terraform is the work of thousands of contributors. We appreciate your help!

To contribute, please read the contribution guidelines: [Contributing to Terraform - vRealize Automation 7 Provider](CONTRIBUTING.md)

Issues on GitHub are intended to be related to bugs or feature requests with provider codebase. See https://www.terraform.io/docs/extend/community/index.html for a list of community resources to ask questions about Terraform.

License
-------

`terraform-provider-vra7` is available under the [Mozilla Public License, version 2.0 license](LICENSE).
