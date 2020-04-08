package vra7

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-vra7/sdk"
	"github.com/terraform-providers/terraform-provider-vra7/utils"
)

func init() {
	fmt.Println("init")
	insecureBool, _ := strconv.ParseBool(mockInsecure)
	client = sdk.NewClient(mockUser, mockPassword, mockTenant, mockBaseURL, insecureBool)
}

func TestConfigValidityFunction(t *testing.T) {

	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	catalogItemID := "dhbh-jhdv-ghdv-dhvdd"

	path := fmt.Sprintf(sdk.RequestTemplateAPI, catalogItemID)
	url := client.BuildEncodedURL(path, nil)

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, mockRequestTemplate))

	resourceConfigList := make([]map[string]interface{}, 0)
	resourceConfigurationObject1 := make(map[string]interface{})

	configMap := make(map[string]interface{})
	configMap["cpu"] = 2
	configMap["memory"] = 1024

	resourceConfigurationObject1["configuration"] = configMap
	resourceConfigurationObject1["component_name"] = "mock.test.machine1"

	resourceConfigList = append(resourceConfigList, resourceConfigurationObject1)

	resourceSchema := resourceVra7Deployment().Schema

	resourceSchemaMap := map[string]interface{}{
		"catalog_item_id":        catalogItemID,
		"wait_timeout":           20,
		"resource_configuration": resourceConfigList,
	}

	mockResourceData := schema.TestResourceDataRaw(t, resourceSchema, resourceSchemaMap)

	p, _ := readProviderConfiguration(mockResourceData, &client)

	mockRequestTemplateStruct, err := p.checkResourceConfigValidity(&client)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, mockRequestTemplateStruct)
	utils.AssertNotNil(t, mockRequestTemplateStruct.Data["mock.test.machine1"])
	utils.AssertEqualsString(t, catalogItemID, mockRequestTemplateStruct.CatalogItemID)

	resourceConfigurationObject2 := make(map[string]interface{})
	resourceConfigurationObject2["configuration"] = configMap
	resourceConfigurationObject2["component_name"] = "mock.test.machine2"

	resourceConfigList = append(resourceConfigList, resourceConfigurationObject2)

	resourceSchemaMap = map[string]interface{}{
		"catalog_item_id":        catalogItemID,
		"resource_configuration": resourceConfigList,
	}

	mockResourceData = schema.TestResourceDataRaw(t, resourceSchema, resourceSchemaMap)
	p, _ = readProviderConfiguration(mockResourceData, &client)

	var mockInvalidKeys []string
	mockInvalidKeys = append(mockInvalidKeys, "mock.test.machine2")

	validityErr := fmt.Sprintf(ConfigInvalidError, strings.Join(mockInvalidKeys, ", "))
	mockRequestTemplateStruct, err = p.checkResourceConfigValidity(&client)
	// this should throw an error mock.test.machine2 does not match
	// with the component names(mock.test.machine1 and machine2) in the request template
	utils.AssertNotNilError(t, err)
	utils.AssertEqualsString(t, validityErr, err.Error())
	utils.AssertNil(t, mockRequestTemplateStruct)
}

func TestAccVra7Deployment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVra7DeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVra7DeploymentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVra7DeploymentExists("vra7_deployment.this"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.this", "description", "Terraform deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.this", "reasons", "Testing the vRA 7 Terraform provider"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.this", "businessgroup_name", "Terraform-BG"),
				),
			},
			{
				Config: testAccCheckVra7DeploymentUpdateConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVra7DeploymentExists("vra7_deployment.this"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.this", "description", "Terraform deployment"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.this", "reasons", "Testing the vRA 7 Terraform provider"),
					resource.TestCheckResourceAttr(
						"vra7_deployment.this", "businessgroup_name", "Terraform-BG"),
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

func testAccCheckVra7DeploymentConfig() string {
	return `
resource "vra7_deployment" "this" {
	catalog_item_name = "Terraform-Simple-BP"
	description = "Terraform deployment"
	reasons = "Testing the vRA 7 Terraform provider"
	lease_days = 20
	deployment_configuration = {
		"BPCustomProperty" = "custom deployment property"
	}
	
	resource_configuration {
		component_name = "vSphere1"
		cluster = 2
		configuration = {
			cpu = 2
			memory = 2048
			vSphere1CustomProperty = "custom machine property"
		}
	}
	
	resource_configuration {
		cluster = 2
		component_name = "vSphere2"
		configuration = {
			cpu = 3
			memory = 2048
		}
	}
	wait_timeout = 20
	businessgroup_name = "Terraform-BG"
}`
}

func testAccCheckVra7DeploymentUpdateConfig() string {
	return `
resource "vra7_deployment" "this" {
	catalog_item_name = "Terraform-Simple-BP"
	description = "Terraform deployment"
	reasons = "Testing the vRA 7 Terraform provider"
	lease_days = 20
	deployment_configuration = {
		"BPCustomProperty" = "custom deployment property"
	}
	
	resource_configuration {
		component_name = "vSphere1"
		cluster = 1
		configuration = {
			cpu = 4
			memory = 2048
			vSphere1CustomProperty = "custom machine property"
		}
	}
	resource_configuration {
		cluster = 1
		component_name = "vSphere2"
		configuration = {
			cpu = 2
			memory = 2048
		}
	}
	wait_timeout = 20
	businessgroup_name = "Terraform-BG"
}`
}
