package vra7

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-vra7/sdk"
	"github.com/terraform-providers/terraform-provider-vra7/utils"
)

func init() {
	fmt.Println("init")
	insecureBool, _ := strconv.ParseBool(mockInsecure)
	client = sdk.NewClient(mockUser, mockPassword, mockTenant, mockBaseURL, insecureBool)
}

func TestConfigValidityFunction(t *testing.T) {

	mockRequestTemplate, err := GetMockRequestTemplate()
	utils.AssertNilError(t, err)

	// a resource_configuration map is created with valid components
	// all combinations of components name and properties are created with dots
	mockConfigResourceMap := make(map[string]interface{})
	mockConfigResourceMap["mock.test.machine1.cpu"] = 2
	mockConfigResourceMap["mock.test.machine1.mock.storage"] = 8

	resourceSchema := resourceVra7Deployment().Schema

	resourceDataMap := map[string]interface{}{
		"catalog_item_id":        "abcdefghijklmn",
		"resource_configuration": mockConfigResourceMap,
	}

	mockResourceData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	p := readProviderConfiguration(mockResourceData)

	readProviderConfiguration(mockResourceData)
	err = p.checkResourceConfigValidity(mockRequestTemplate)
	utils.AssertNilError(t, err)

	mockConfigResourceMap["machine2.mock.cpu"] = 2
	mockConfigResourceMap["machine2.storage"] = 2

	resourceDataMap = map[string]interface{}{
		"catalog_item_id":        "abcdefghijklmn",
		"resource_configuration": mockConfigResourceMap,
	}

	mockResourceData = schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	readProviderConfiguration(mockResourceData)

	err = p.checkResourceConfigValidity(mockRequestTemplate)
	utils.AssertNilError(t, err)

	mockConfigResourceMap["mock.machine3.vSphere.mock.cpu"] = 2
	resourceDataMap = map[string]interface{}{
		"catalog_item_id":        "abcdefghijklmn",
		"resource_configuration": mockConfigResourceMap,
	}

	mockResourceData = schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	p = readProviderConfiguration(mockResourceData)

	var mockInvalidKeys []string
	mockInvalidKeys = append(mockInvalidKeys, "mock.machine3.vSphere.mock.cpu")

	validityErr := fmt.Sprintf(ConfigInvalidError, strings.Join(mockInvalidKeys, ", "))
	err = p.checkResourceConfigValidity(mockRequestTemplate)
	// this should throw an error as none of the string combinations (mock, mock.machine3, mock.machine3.vsphere, etc)
	// matches the component names(mock.test.machine1 and machine2) in the request template
	utils.AssertNotNilError(t, err)
	utils.AssertEqualsString(t, validityErr, err.Error())
}

// creates a mock request template from a request template template json file
func GetMockRequestTemplate() (*sdk.CatalogItemRequestTemplate, error) {

	mockRequestTemplateStruct := sdk.CatalogItemRequestTemplate{}
	err := json.Unmarshal([]byte(mockRequestTemplate), &mockRequestTemplateStruct)
	if err != nil {
		return nil, err
	}

	return &mockRequestTemplateStruct, nil

}

func TestAccVra7DeploymentCreate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVra7DeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVra7DeploymentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVra7DeploymentExists("vra7_deployment.my_vra7_deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "description", "Test deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "reasons", "Testing the vRA 7 Terraform plugin"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "businessgroup_name", "Development"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration.%", "2"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration._leaseDays", "15"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration.deployment_property", "custom deployment property"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.%", "3"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.cpu", "1"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.memory", "2048"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.machine_property", "machine custom property"),
				),
			},
		},
	})
}

func TestAccVra7DeploymentUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVra7DeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVra7DeploymentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVra7DeploymentExists("vra7_deployment.my_vra7_deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "description", "Test deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "reasons", "Testing the vRA 7 Terraform plugin"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "businessgroup_name", "Development"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration.%", "2"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration._leaseDays", "15"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration.deployment_property", "custom deployment property"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.%", "3"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.cpu", "1"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.memory", "2048"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.machine_property", "machine custom property"),
				),
			},
			{
				Config: testAccCheckVra7DeploymentUpdateConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVra7DeploymentExists("vra7_deployment.my_vra7_deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "description", "Test deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "reasons", "Testing the vRA 7 Terraform plugin"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "businessgroup_name", "Development"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration.%", "2"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration._leaseDays", "15"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "deployment_configuration.deployment_property", "updated custom deployment property"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.%", "3"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.cpu", "1"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.memory", "1024"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.my_vra7_deployment", "resource_configuration.vSphereVM1.machine_property", "updated machine custom property"),
				),
			},
		},
	})
}

func testAccCheckVra7DeploymentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource request ID is set")
		}

		return nil
	}
}

func testAccCheckVra7DeploymentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vra7_deployment" {
			continue
		}

		_, err := client.GetRequestResourceView(rs.Primary.ID)
		if err == nil {
			return err
		}
	}
	return nil
}

func testAccCheckVra7DeploymentUpdateConfig() string {
	return fmt.Sprintf(`
resource "vra7_deployment" "my_vra7_deployment" {
count = 1
catalog_item_name = "Basic Single Machine"
description = "Test deployment"
reasons = "Testing the vRA 7 Terraform plugin"

deployment_configuration = {
	_leaseDays = "15"
	deployment_property = "updated custom deployment property"
}
resource_configuration = {
	"vSphereVM1.cpu" = 1
	"vSphereVM1.memory" = 1024
	"vSphereVM1.machine_property" = "updated machine custom property"
}
wait_timeout = 20
businessgroup_name = "Development"
}`)
}

func testAccCheckVra7DeploymentConfig() string {
	return fmt.Sprintf(`
resource "vra7_deployment" "my_vra7_deployment" {
count = 1
catalog_item_name = "Basic Single Machine"
description = "Test deployment"
reasons = "Testing the vRA 7 Terraform plugin"

deployment_configuration = {
	_leaseDays = "15"
	deployment_property = "custom deployment property"
}
resource_configuration = {
	"vSphereVM1.cpu" = 1
	"vSphereVM1.memory" = 2048
	"vSphereVM1.machine_property" = "machine custom property"
}
wait_timeout = 20
businessgroup_name = "Development"
}`)
}
