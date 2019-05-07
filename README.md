VMware Terraform provider for vRealize Automation 7
===
[![Build Status](https://travis-ci.org/vmware/terraform-provider-vra7.svg?branch=master)](https://travis-ci.org/vmware/terraform-provider-vra7)

A self-contained deployable integration between Terraform and vRealize Automation (vRA) which allows Terraform users to request/provision entitled vRA catalog items using Terraform. Supports Terraform destroying vRA provisioned resources.

# Getting Started
These instructions will get you a copy of the project up and run on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

# Prerequisites
To get the vRA plugin up and running you need the following things.
* [Terraform 0.9 or above](https://www.terraform.io/downloads.html)
* [Go Language 1.11.4 or above](https://golang.org/dl/)

# Project Setup
Setup a GOLang project structure as follows:
```
|-/home/<USER>/TerraformPluginProject/
    |-bin/
    |-pkg/
    |-src/

```

# Environment Setup
Set following environment variables.

## Linux Users
`GOROOT` is a golang library path.
```
export GOROOT=/usr/local/go
```

`GOPATH` is a path pointing toward the source code directory.
```
export GOPATH=/home/<user>/TerraformPluginProject
```

**Windows Users**

*GOROOT is a golang library path*
```
set GOROOT=C:\Go
```

`GOPATH` is a path pointing toward the source code directory.
```
set GOPATH=C:\TerraformPluginProject
```

## Set terraform provider

## Linux Users
Create `~/.terraformrc` and put following content in it.
```
    providers {
         vra7 = "/home/<USER>/TerraformPluginProject/bin/terraform-provider-vra7"
    }
```

## Windows Users
Create `%APPDATA%/terraform.rc` and put following content in it.
```
    providers {
         vra7 = "C:\\TerraformPluginProject\\bin\\terraform-provider-vra7.exe"
    }
```


# Installation
Clone repo code into go project using *go get*
```
    go get github.com/terraform-providers/terraform-provider-vra7
```

# Create Binary

## Linux and MacOS Users
Navigate to `/home/<USER>/TerraformPluginProject/src/github.com/terraform-providers/terraform-provider-vra7` and run go build command to generate plugin binary.
```
    go build -o /home/<USER>/TerraformPluginProject/bin/terraform-provider-vra7

```

## Windows Users
Navigate to `C:\TerraformPluginProject\src\github.com\vmware\terraform-provider-vra7` and run go build command to generate plugin binary.
```
    go build -o C:\TerraformPluginProject\bin\terraform-provider-vra7.exe

```

# Usage

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
         Linux.cpu = "1"
         Windows2008R2SP1.cpu =  "2"
         Windows2012.cpu =  "4"
         Windows2016.cpu =  "2"
     }
     deployment_configuration = {
         _leaseDays = "5"
     }
     count = 3
}
```

**Example 2:**
```
resource "vra7_deployment" "example_machine2" {
  catalog_item_id = "e5dd4fba7f96239286be45ed"
   resource_configuration = {
         Linux.cpu = "1"
         Windows2008.cpu =  "2"
         Windows2012.cpu =  "4"
         Windows2016.cpu =  "2"
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

# Scripts

**update_resource_state.sh**

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

# Contributing
The `terraform-provider-vra7` project team welcomes contributions from the community. Before you start working with `terraform-provider-vra7`, please read our [Developer Certificate of Origin](https://cla.vmware.com/dco). All contributions to this repository must be signed as described on that page. Your signature certifies that you wrote the patch or have the right to pass it on as an open-source patch. For more detailed information, refer to [CONTRIBUTING.md](CONTRIBUTING.md).

# License
`terraform-provider-vra7` is available under the [Mozilla Public License, version 2.0 license](LICENSE).
