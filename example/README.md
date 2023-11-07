# VMware vRA Terraform Plugin

A self-contained deployable integration between Terraform and vRealize Automation (vRA) which allows Terraform users to request/provision entitled vRA catalog items using Terraform. Supports Terraform destroying vRA provisioned resources.

Examples:
- simple - simple single machine example
- depends_on - example illustrating how to make one machine depend on another
- multi-machine - example showing multiple machines in one catalog
- remote-execute - example showing how to remote-execute commands on a VM as soon as it's created


## Prerequisites

Follow the steps mentioned in Main [README](../README.md)

## How to run the example
1. cd <example directory>
2. Copy the `terraform.tfvars.sample` to `terraform.tfvars`
3. Copy the created plugin binary to example/ (optional)
4. Edit the `terraform.tfvars` to add the secret information
5. To init command -  `terraform init`
6. To plan command - `terraform plan`
7. To apply command - `terraform apply`
8. To destroy command - `terraform destroy`
