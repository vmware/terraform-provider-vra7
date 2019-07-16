---
layout: "vra7"
page_title: "Provider: VMware vRA7"
sidebar_current: "docs-vra7-index"
description: |-
  A Terraform provider to work with VMware vRealize Automation 7, allowing deployment of existing blueprints.
---

# VMware vRA7 Provider

The VMware vRA7 provider gives Terraform the ability to work with [VMware vRealize
Automation 7][vmware-vra]. This provider can be used to deploy exisiting blueprints
as deployments.

[vmware-vra]: https://www.vmware.com/products/vrealize-automation.html

Use the navigation on the left to read about the various resources and data
sources supported by the provider.

## Example Usage

The following abridged example demonstrates a current basic usage of the
provider to launch a virtual machine using the [`vra7_deployment`
resource][tf-vra7-deployment]. 

[tf-vra7-deployment]: /docs/providers/vra7/r/deployment.html

```hcl
provider "vra7" {
  username = "${var.username}"
  password = "${var.password}"
  tenant   = "${var.tenant}"
  host     = "${var.host}"
}

resource "vra7_deployment" "machine" {
  count             = 1
  catalog_item_name = "CentOS 7.0 x64"

  resource_configuration = {
    "Linux.cpu" = "2"
  }
}
```

See the sidebar for usage information on all of the resources, which will have
examples specific to their own use cases.

## Argument Reference

The following arguments are used to configure the VMware vRA7 Provider:

* `username` - (Required) This is the username for vRA7 API operations. Can also
  be specified with the `VRA7_USERNAME` environment variable.
* `password` - (Required) This is the password for vRA7 API operations. Can
  also be specified with the `VRA7_PASSWORD` environment variable.
* `tenant` - (Required) This is the vRA tenant ID vRA API
  operations. Can also be specified with the `VRA7_SERVER` environment
  variable.
* `host` - (Required) This is the vRA server name for vRA API
  operations. Can also be specified with the `VRA7_HOST` environment
  variable.
* `insecure` - (Optional) Boolean that can be set to true to
  disable SSL certificate verification. This should be used with care as it
  could allow an attacker to intercept your auth token. If omitted, default
  value is `false`. Can also be specified with the `VRA7_INSECURE`
  environment variable.

### Debugging options

~> **NOTE:** The following options can leak sensitive data and should only be
enabled when instructed to do so by HashiCorp for the purposes of
troubleshooting issues with the provider, or when attempting to perform your
own troubleshooting. Use them at your own risk and do not leave them enabled!

## Bug Reports and Contributing

For more information how how to submit bug reports, feature requests, or
details on how to make your own contributions to the provider, see the vRA7
provider [project page][tf-vra7-project-page].

[tf-vra7-project-page]: https://github.com/terraform-providers/terraform-provider-vra7


