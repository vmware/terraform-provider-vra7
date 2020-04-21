## 1.1.0 (Unreleased)
## 1.0.1 (April 21, 2020)

BUG FIXES:
* Terraform crash : rpc error: code = Unavailable desc = transport is closing ([#59](https://github.com/terraform-providers/terraform-provider-vra7/issues/59))
* Provider 1.0.0 - Required to specify all values even if there's a default specified in the blueprint ([#60](https://github.com/terraform-providers/terraform-provider-vra7/issues/60))

## 1.0.0 (April 08, 2020)

BUG FIXES:
* Terraform gets only first VM with _number_of_instances or _cluster > 1  ([#39](https://github.com/terraform-providers/terraform-provider-vra7/issues/39))
* Terraform refresh does not work as intended ([#38](https://github.com/terraform-providers/terraform-provider-vra7/issues/38))
* VRA Provider deletes resources from state file before receiving any "SUCCESSFUL" status response from VRA during terraform destroy ([#33](https://github.com/terraform-providers/terraform-provider-vra7/issues/33))
* Terraform apply returns empty ip address ([#27](https://github.com/terraform-providers/terraform-provider-vra7/issues/27))
* IPAddress return is empty with latest release of vRealize Automation ([#16](https://github.com/terraform-providers/terraform-provider-vra7/issues/16))
* Terraform adds resource in state file even though the request_status is "FAILED" ([#37](https://github.com/terraform-providers/terraform-provider-vra7/issues/37))

FEATURES:
* The resource schema is modified to have more attributes for the vra7_deployment resource
* Feature request: Day 2: Change the number of VMs created by the vRA blueprint ([#47](https://github.com/terraform-providers/terraform-provider-vra7/issues/47))
* Does 'import' work? ([#29](https://github.com/terraform-providers/terraform-provider-vra7/issues/29))
* Need to import existing VMs ([#43](https://github.com/terraform-providers/terraform-provider-vra7/issues/43))
* Support Deployment Day 2 Change Lease action ([#54](https://github.com/terraform-providers/terraform-provider-vra7/issues/54))
* Create a data source for vra7_deployment resource ([#55](https://github.com/terraform-providers/terraform-provider-vra7/issues/55))

IMPROVEMENTS:
* Cleanup README
* Show all the data in the state file that is returned from a deployment resource GET ([#41](https://github.com/terraform-providers/terraform-provider-vra7/issues/41))
* Cannot pass array of values in element of deployment_configuration or resource_configuration ([#45](https://github.com/terraform-providers/terraform-provider-vra7/issues/45))
* `ip_address` can be accessed from the resource_configuration schema as a first class attribute


## 0.5.0 (November 06, 2019)
FEATURES:

IMPROVEMENTS:

BUG FIXES:

* Added logic to pull network info from NETWORK_LIST json in resourceview

## 0.4.1 (August 08, 2019)

* 0.4.0 missed 0.12 support which this release aims to ship with.

## 0.4.0 (August 08, 2019)
FEATURES:

IMPROVEMENTS:

* Upgrade terraform SDK code to v0.12.6 ([#26](https://github.com/terraform-providers/terraform-provider-vra7/pull/26))

BUG FIXES:

## 0.3.0 (August 07, 2019)

FEATURES:

IMPROVEMENTS:

* Updates for terraform 0.12.0. Thanks @skylerto! ([#17](https://github.com/terraform-providers/terraform-provider-vra7/pull/17))
* Changes to make acceptance tests run with v0.12 changes ([#18](https://github.com/terraform-providers/terraform-provider-vra7/pull/18))
* Updates to examples to match return types ([#21](https://github.com/terraform-providers/terraform-provider-vra7/pull/21))

BUG FIXES:

* Cleanup README
* Fix travis tests and changes to pass linting ([#10](https://github.com/terraform-providers/terraform-provider-vra7/pull/10))
* Formatting example code and removing debugging comment ([#11](https://github.com/terraform-providers/terraform-provider-vra7/pull/11))
* Update failure was returning wrong status in the console. ([#22](https://github.com/terraform-providers/terraform-provider-vra7/pull/22))
* The provider should wait for the terminal state or timeout ([#24](https://github.com/terraform-providers/terraform-provider-vra7/pull/24))


## 0.2.0 (May 07, 2019)

FEATURES:

IMPROVEMENTS:

* Rename dirs/files according to the hashicorp provider's guidelines ([#145](https://github.com/vmware/terraform-provider-vra7/pull/145))

* Acceptance tests for vra7_deployment resource and fix for issue # 143 ([#144](https://github.com/vmware/terraform-provider-vra7/pull/144))


BUG FIXES:

* Acceptance tests for vra7_deployment resource and fix for issue # 143 ([#144](https://github.com/vmware/terraform-provider-vra7/pull/144))


## 0.1.0 (April 1, 2019)

FEATURES:

IMPROVEMENTS:

* Changes in the tf config file schema ([#135](https://github.com/vmware/terraform-provider-vra7/pull/135))

BUG FIXES:


## 0.0.2 (March 26, 2019)

FEATURES:

IMPROVEMENTS:

* Refactor code to split provider and SDK ([#119](https://github.com/vmware/terraform-provider-vra7/pull/119))

* Add more unit tests for the sdk and some refactoring ([#128](https://github.com/vmware/terraform-provider-vra7/pull/128))

BUG FIXES:

* Handle response pagination when fetching catalog item id by name ([#134](https://github.com/vmware/terraform-provider-vra7/pull/134))


## 0.0.1 (February 7, 2019)

FEATURES:

* Add requirement for go 1.11.4 or above ([#122](https://github.com/vmware/terraform-provider-vra7/issues/122))
* Convert from using dep to go modules ([#109](https://github.com/vmware/terraform-provider-vra7/issues/109))
* Adding businessgroup_name in the config file ([#94](https://github.com/vmware/terraform-provider-vra7/issues/94))
* Adding code to check if the component names in the terraform resource… ([#88](https://github.com/vmware/terraform-provider-vra7/issues/88))
* Get VM IP address ([#66](https://github.com/vmware/terraform-provider-vra7/issues/66))
* Update Deployment based on changes to configuration in Terraform file ([#27](https://github.com/vmware/terraform-provider-vra7/issues/27))
* resource_configuration key format verification check ([#36](https://github.com/vmware/terraform-provider-vra7/issues/36))
* Business Group Id resource field ([#28](https://github.com/vmware/terraform-provider-vra7/issues/28))
* Initial Pass at allowing 'description' and 'reasons' to be specified for a deployment ([#12](https://github.com/vmware/terraform-provider-vra7/issues/12))
* #7 Terraform "depends_on" does not wait - wait_timeout resource schema added. ([#10](https://github.com/vmware/terraform-provider-vra7/issues/10))
* Add insecure setting to allow connection with self-signed certs ([#3](https://github.com/vmware/terraform-provider-vra7/issues/3))

IMPROVEMENTS:

* Update README.md ([#114](https://github.com/vmware/terraform-provider-vra7/issues/114))
* Adding a logging framework for more detailed logging of vRA Terraform plugging. ([#85](https://github.com/vmware/terraform-provider-vra7/issues/85))
* Added debug messages to resource.go to help debug issues in the field. ([#80](https://github.com/vmware/terraform-provider-vra7/issues/80))
* Changes to variable and function names to better reflect vRA terminology ([#65](https://github.com/vmware/terraform-provider-vra7/issues/65))
* README.md changes ([#62](https://github.com/vmware/terraform-provider-vra7/issues/62))
* Unit testing - code coverage ([#48](https://github.com/vmware/terraform-provider-vra7/issues/48))
* Clean up the resource section of the README ([#32](https://github.com/vmware/terraform-provider-vra7/issues/32))
* Certificate signed by unknown authority README updates ([#16](https://github.com/vmware/terraform-provider-vra7/issues/16))
* Multi-machine blueprint terraform config example. ([#13](https://github.com/vmware/terraform-provider-vra7/issues/13))

BUG FIXES:

* Update go sum to fix the build failure ([#121](https://github.com/vmware/terraform-provider-vra7/issues/121))
* lease_days property name should be _leaseDays. ([#112](https://github.com/vmware/terraform-provider-vra7/issues/112))
* Have golint errors fail "make check" ([#108](https://github.com/vmware/terraform-provider-vra7/issues/108))
* Fix go lint errors/warnings ([#106](https://github.com/vmware/terraform-provider-vra7/issues/106))
* Cleanup travis tests ([#105](https://github.com/vmware/terraform-provider-vra7/issues/105))
* Fix terraform destroy. ([#103](https://github.com/vmware/terraform-provider-vra7/issues/103))
* Update issue templates ([#102](https://github.com/vmware/terraform-provider-vra7/issues/102))
* Correction in the schema ([#99](https://github.com/vmware/terraform-provider-vra7/issues/99))
* Fixing issues related to create, update and read ([#98](https://github.com/vmware/terraform-provider-vra7/issues/98))
* Fixing the config validation bug # 91 ([#92](https://github.com/vmware/terraform-provider-vra7/issues/92))
* Show request status on terraform update operation ([#90](https://github.com/vmware/terraform-provider-vra7/issues/90))
* Updating the request_status properly on time out. ([#86](https://github.com/vmware/terraform-provider-vra7/issues/86))
* merge crash fixes - minor change in add new value to machine config ([#64](https://github.com/vmware/terraform-provider-vra7/issues/64))
* Changes to the resource creation flow ([#55](https://github.com/vmware/terraform-provider-vra7/issues/55))
* Issue fix : Terraform destroy runs async (completes immediately) ([#56](https://github.com/vmware/terraform-provider-vra7/issues/56))
* Redo the error login in deleteResource to prevent panic ([#38](https://github.com/vmware/terraform-provider-vra7/issues/38))
* Add dynamic/deploy time properties appropriately from resource_configuration block ([#25](https://github.com/vmware/terraform-provider-vra7/issues/25))
* Use SplitN instead of Split to identify fields to replaces ([#29](https://github.com/vmware/terraform-provider-vra7/issues/29))
* Corrected minor typos in README.md ([#30](https://github.com/vmware/terraform-provider-vra7/issues/30))
* destroy resource outside terraform  error message fixes ([#22](https://github.com/vmware/terraform-provider-vra7/issues/22))
