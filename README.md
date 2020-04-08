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
* [vRealize Automation 7.5 or above](https://www.vmware.com/products/vrealize-automation.html)

Using the provider
----------------------

See the [vra7 documentation](https://www.terraform.io/docs/providers/vra7/index.html) to get started using the vRealize Automation 7 provider.

See vra7_deployment resource examples [here] (examples/README.md)

## Execution
These are the Terraform commands that can be used for the vRA plugin:
* `terraform init` - The init command is used to initialize a working directory containing Terraform configuration files.
* `terraform plan` - Plan command shows plan for resources like how many resources will be provisioned and how many will be destroyed.
* `terraform apply` - apply is responsible to execute actual calls to provision resources.
* `terraform refresh` - By using the refresh command you can check the status of the request.
* `terraform show` - show will set a console output for resource configuration and request status.
* `terraform destroy` - destroy command will destroy all the  resources present in terraform configuration file.

Navigate to the location where `main.tf` and binary are placed and use the above commands as needed.

Upgrading the provider
----------------------

The vra7 provider doesn't upgrade automatically once you've started using it. After a new release you can run 

```bash
terraform init -upgrade
```

to upgrade to the latest stable version of the vra7 provider. See the [Terraform website](https://www.terraform.io/docs/configuration/providers.html#provider-versions)
for more information on provider upgrades, and how to set version constraints on your provider.

Migrating from previous versions to version 1.0.0, issues fixed and enhancements
---------------------------------------------------------------------------------

There are some schema changes in the provider version 1.0.0. These changes are made to support vra7 deployment Day 2 actions, detailed information of the deployment in the state file, getting access to more deployment and resource level properties, esp. `ip_address`, etc. See the release notes for more detail.

### Previous main.tf file

```hcl
provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

resource "vra7_deployment" "this" {
    count                      = 1
    catalog_item_name          = "multi_machine_catalog"
    businessgroup_name         = Development
    wait_timeout               = 20
    deployment_configuration = {
        "_leaseDays"                 = "15"                   //number of lease days
        "BPCustomProp"               = "custom depl prop"     //custom property in BP required while requesting a catalog item
        "Container"                  = "App.Container"        //property of a property group
        "Container.Auth.User"        = "var.container_user"   //property of a property group
        "Container.Auth.Password"    = "var.container_pw"     //property of a property group
        "Container.Connection.Port"  = "var.container_port"   //property of a property group
    }

    resource_configuration = {
        "Windows.cpu"            = "2"                //Windows Machine CPU
        "Windows.memory"         = "1024"             //Windows Machine memory
        "Windows.vm_custom_prop" = "a custom prop"    //Windows custom property called vm_custom_property
        "Windows._cluster"       = "2"                //Windows cluster size
        "Linux.cpu"              = "2"                //Linux Machine CPU
        "http.hostname"          = "xyz.com"          //HTTP (apache) hostname
        "http.network_mode"      = "bridge"           //HTTP (apache) network mode
    }
}
```
Migrating to the latest version would require to make the following changes in the TF confile(main.tf).

* `_leaseDays` is moved out of of deployment_configuration and added as a property in the schema. The name of the property is `lease_days`. This can be changed for Change Lease Day 2 action.
* We create a `resource_configuration` block for each component. There are three components, Windows, Linux and http.
For instance, the resource_configuration for Windows component would look like this:

```hcl
resource_configuration {
    component_name              = "Windows"          //This is the component name and need not be prefixed with all properties
    cluster                     = 2                  //cluster is added as a property in the schema. Modifying it will do the
                                                     // Scale In/Out Day 2 actions
    configuration = {
        cpu                     = 2
        memory                  = 1024
        vm_custom_prop          = "a custom prop"
    }
}
```
* `_cluster` is added as a property in the schema. It can be modified for Scale In/Scale Out Day 2 actions
* Support for deployemnt and resource properties of type array of strings in the blocks deployment_configuration as well as configuration under resource_configuration respectively as shown in the example below.
* `ip_address` need not be added in the main.tf. It can be accessed from the state file. It is added as a property in the resource_configuration schema. Please refer to the documentation.

### The new main.tf file would look as follows:

```hcl
provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

resource "vra7_deployment" "this" {
    count                      = 1
    catalog_item_name          = "multi_machine_catalog"
    businessgroup_name         = Development
    wait_timeout               = 20
    lease_days                 = 15                           //number of lease days

    deployment_configuration = {
        "BPCustomProp"               = "custom depl prop"     //custom property in BP required while requesting a catalog item
        "Container"                  = "App.Container"        //property of a property group
        "Container.Auth.User"        = "var.container_user"   //property of a property group
        "Container.Auth.Password"    = "var.container_pw"     //property of a property group
        "Container.Connection.Port"  = "var.container_port"   //property of a property group
        "businessGroups" = <<EOF                              //this is an example to property of type array of strings
        [
            "bgTest1",
            "bgTest2"
        ]
        EOF
    }

    resource_configuration {
        component_name              = "Windows"       //This is the component name and need not be prefixed with all properties
        cluster                     = 2               //cluster is added as a property in the schema. Modifying it will do the
                                                      // Scale In/Out Day 2 actions
        configuration = {
            cpu                     = 2
            memory                  = 1024
            vm_custom_prop          = "a custom prop"
        }
    }

    resource_configuration {
        component_name              = "Linux"      //This is the component name and need not be prefixed with all properties
        configuration = {
            cpu                     = 2
            security_tag = <<EOF                   //this is an example to property of type array of strings
            [
                "dev_sg",
                "prod_sg"
            ]
            EOF
        }
    }

    resource_configuration {
        component_name              = "http"      //This is the component name and need not be prefixed with all properties
        configuration = {
            hostname                = "xyz.com"          //HTTP (apache) hostname
            network_mode            = "bridge"           //HTTP (apache) network mode
        }
    }
}
```

## Import vra7_deployment

Import functionality is now supported for the vra7_deployment resource. If there is an exiting deployment, it can be imported by catalog item request id.

### main.tf

```hcl
provider "vra7" {
  username = var.username
  password = var.password
  tenant   = var.tenant
  host     = var.host
}

resource vra7_deployment "this" {
    // the properties can be added once the import is completed by referring to the state file
}
```
terraform import vra7_deployment.this <request_id>

## Data source vra7_deployment

A data source for vra7_deployment can also be created using either deployment ID or catalog item request id.
Refer to the documentation [here](website/docs/d/vra7_deployment.html.markdown)


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

Contributing
------------

Terraform is the work of thousands of contributors. We appreciate your help!

To contribute, please read the contribution guidelines: [Contributing to Terraform - vRealize Automation 7 Provider](CONTRIBUTING.md)

Issues on GitHub are intended to be related to bugs or feature requests with provider codebase. See https://www.terraform.io/docs/extend/community/index.html for a list of community resources to ask questions about Terraform.

License
-------

`terraform-provider-vra7` is available under the [Mozilla Public License, version 2.0 license](LICENSE).
