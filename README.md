Terraform provider for VMware vRealize Automation 7
==================

> **_NOTE:_** The [End of General Support for VMware vRealize Suite7.x](https://knowledge.broadcom.com/external/article/326048/end-of-general-support-for-vmware-vreali.html) (which includes vRA 7.x, vRops 7.x and vRLI 7.x) was on September 1st, 2022. Therefore, VMware has also ended the active development of this Terraform Provider, so this repository will no longer be updated. We recommend users migrate to the [VMware vRealize Automation v8 Provider](https://github.com/vmware/terraform-provider-vra).

- Website: https://www.terraform.io
- Documentation: https://www.terraform.io/docs/providers/vra7/index.html

Introduction
------------

A self-contained deployable integration between Terraform and vRealize Automation (vRA) which allows Terraform users to request/provision entitled vRA catalog items using Terraform. Supports Terraform destroying vRA provisioned resources.

Requirements
------------

* [Terraform 0.9 or above](https://www.terraform.io/downloads.html)
* [Go Language 1.11.4 or above](https://golang.org/dl/)
* [vRealize Automation 7.4 or below] The support has been stopped since provider v3.0.0. It is recommended to use the previous versions of the provider (v2.0.1 or below)
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


### A sample main.tf file is as follows:

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

## Supported Day 2 actions. Examples are provided in the documentation

1. Change lease: To do this, add the new expiry_date in the config file.
2. Scale: Change the cluster size
3. Reconfigure: Modify/add properties inside the configuration block under resource_configuration block

## Outputs

The resource_configuration block has an instances block that is a list of all the instances/VMs corresponding to a component. The instance list size is nothing but the custer size.

For example, after the deployment is created using the above config file, the resource_configuration list size will be 3.
And the instances list size in the resource configuration map corresonding to the component "Windows" will be 2. This is because the cluster size is 2 and it creates 2 VMs with that configuration.

Sample outputs:

```
output "ip_address" {
    value = vra7_deployment.this[*].resource_configuration[*].instances[*].properties.ip_address
}
```
```
output "component" {
    value = vra7_deployment.this[*].resource_configuration[*].component_name
}
```

```
output "vm_name" {
    value = vra7_deployment.this[*].resource_configuration[*].instances[*].properties.name
}
```

Expected sample outputs (based on the above main.tf, ip_address and vm_names are mock data below):

```
ip_address = [
  [
    [
      "10.xxx.xxx.xxx",
      "10.xxx.xxx.xxx",
    ],
    [
      "10.xxx.xxx.xxx"
    ],
    [
      "10.xxx.xxx.xxx"
    ],
  ],
]


component = [
  [
    "Windows",
    "Linux",
    "http",
  ],
]

vm_name = [
  [
    [
      "Windows-machine1-2048",
      "Windows-machine2-2049",
    ],
    [
      "Linux-machine1-2050",
    ],
    [
      "http-machine1-2051",
    ],
  ],
]
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

Clone repository to: `$GOPATH/src/github.com/vmware/terraform-provider-vra7`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-vra7
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/vmware/terraform-provider-vra7
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
