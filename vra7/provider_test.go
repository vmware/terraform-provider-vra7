package vra7

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-vra7/sdk"
	"github.com/terraform-providers/terraform-provider-vra7/utils"
	"gopkg.in/jarcoal/httpmock.v1"
	"os"
	"strconv"
	"testing"
)

var (
	client           sdk.APIClient
	mockUser         = os.Getenv("VRA7_USERNAME")
	mockPassword     = os.Getenv("VRA7_PASSWORD")
	mockTenant       = os.Getenv("VRA7_TENANT")
	mockBaseURL      = os.Getenv("VRA7_HOST")
	mockInsecure     = os.Getenv("VRA7_INSECURE")
	testAccProviders map[string]terraform.ResourceProvider
	testAccProvider  *schema.Provider
)

func init() {

	fmt.Println("init")
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"vra7": testAccProvider,
	}
	insecureBool, _ := strconv.ParseBool(mockInsecure)
	client = sdk.NewClient(mockUser, mockPassword, mockTenant, mockBaseURL, insecureBool)
}

func TestValidateProvider(t *testing.T) {
	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/identity/api/tokens", mockBaseURL),
		httpmock.NewStringResponder(200, validAuthResponse))

	err := client.Authenticate()
	utils.AssertNilError(t, err)

	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/identity/api/tokens", mockBaseURL),
		httpmock.NewStringResponder(90135, errorAuthResponse))

	err = client.Authenticate()
	utils.AssertNotNilError(t, err)
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {

	if v := os.Getenv("VRA7_HOST"); v == "" {
		t.Fatal("VRA7_HOST must be set for acceptance tests")
	}

	if os.Getenv("VRA7_USERNAME") == "" {
		t.Fatal("VRA7_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("VRA7_PASSWORD"); v == "" {
		t.Fatal("VRA7_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("VRA7_TENANT"); v == "" {
		t.Fatal("VRA7_TENANT must be set for acceptance tests")
	}

	if v := os.Getenv("VRA7_INSECURE"); v == "" {
		t.Fatal("VRA7_INSECURE must be set for acceptance tests")
	}
}
