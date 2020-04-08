package sdk

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-vra7/utils"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

var (
	client       APIClient
	mockUser     = os.Getenv("VRA7_USERNAME")
	mockPassword = os.Getenv("VRA7_PASSWORD")
	mockTenant   = os.Getenv("VRA7_TENANT")
	mockBaseURL  = os.Getenv("VRA7_HOST")
	mockInsecure = os.Getenv("VRA7_INSECURE")
)

func init() {
	fmt.Println("init")
	insecureBool, _ := strconv.ParseBool(mockInsecure)
	client = NewClient(mockUser, mockPassword, mockTenant, mockBaseURL, insecureBool)
}

func TestGetCatalogItemRequestTemplate(t *testing.T) {
	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	catalogItemID := "feaedf73-560c-4612-a573-41667e017691"

	path := fmt.Sprintf(RequestTemplateAPI, catalogItemID)
	url := client.BuildEncodedURL(path, nil)

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, requestTemplateResponse))

	catalogItemReqTemplate, err := client.GetCatalogItemRequestTemplate(catalogItemID)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, catalogItemReqTemplate)
	utils.AssertEqualsString(t, catalogItemID, catalogItemReqTemplate.CatalogItemID)

	catalogItemReqTemplate, err = client.GetCatalogItemRequestTemplate("635e5v-8e37efd60-hdgdh")
	utils.AssertNotNilError(t, err)

	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(20116, requestTemplateErrorResponse))
	invalidCatalogItemID := "feaedf73-560c-4612-a573-0041667e0176"
	catalogItemReqTemplate, err = client.GetCatalogItemRequestTemplate(invalidCatalogItemID)
	utils.AssertNotNilError(t, err)
	utils.AssertNil(t, catalogItemReqTemplate)

}

func TestReadCatalogItemNameByID(t *testing.T) {

	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	catalogItemID := "e5dd4fba-45ed-4943-b1fc-7f96239286be"
	path := fmt.Sprintf(EntitledCatalogItems+"/"+"%s", catalogItemID)
	url := client.BuildEncodedURL(path, nil)

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, catalogItemResp))

	catalogItemName, err := client.ReadCatalogItemNameByID(catalogItemID)
	utils.AssertNilError(t, err)
	utils.AssertEqualsString(t, "CentOS 6.3", catalogItemName)

	catalogItemName, err = client.ReadCatalogItemNameByID("84rg=73dv-dd8dhy-hg")
	utils.AssertNotNilError(t, err)
	utils.AssertEqualsString(t, "", catalogItemName)
}

func TestReadCatalogItemByName(t *testing.T) {
	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	url := client.BuildEncodedURL(EntitledCatalogItemViewsAPI, map[string]string{
		"page": "1"})

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, entitledCatalogItemViewsResponse))

	catalogItemID, err := client.ReadCatalogItemByName("CentOs")
	utils.AssertEqualsString(t, "feaedf73-560c-4612-a573-41667e017691", catalogItemID)
	utils.AssertNilError(t, err)

	catalogItemID, err = client.ReadCatalogItemByName("Invalid Catalog Item name")
	utils.AssertEqualsString(t, "", catalogItemID)
}

func TestGetBusinessGroupID(t *testing.T) {

	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	path := Tenants + "/" + mockTenant + "/subtenants"
	url := client.BuildEncodedURL(path, nil)

	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, subTenantsResponse))

	id, err := client.GetBusinessGroupID("Development", mockTenant)
	utils.AssertNilError(t, err)
	utils.AssertEqualsString(t, "b2470b94-cbca-43db-be37-803cca7b0f1a", id)
}

func TestGetRequestStatus(t *testing.T) {
	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	mockRequestID := "adca9535-4a35-4981-8864-28643bd990b0"
	path := fmt.Sprintf(ConsumerRequests+"/"+"%s", mockRequestID)
	url := client.BuildEncodedURL(path, nil)
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, requestStatusResponse))

	requestStatus, err := client.GetRequestStatus(mockRequestID)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, requestStatus)
	utils.AssertEqualsString(t, "IN_PROGRESS", requestStatus.Phase)

	// invalid request id
	mockRequestID = "gd78tegd-0e737egd-jhdg"
	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(20111, requestStatusErrResponse))
	requestStatus, err = client.GetRequestStatus(mockRequestID)
	utils.AssertNotNilError(t, err)
	utils.AssertNil(t, requestStatus)
}

func TestGetRequestResourceView(t *testing.T) {
	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	mockRequestID := "594bf7ec-c8d2-4a0d-8477-553ed987aa48"
	path := fmt.Sprintf(GetRequestResourceViewAPI, mockRequestID)
	url := client.BuildEncodedURL(path, nil)
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, deploymentStateResponse))

	resourceView, err := client.GetRequestResourceView(mockRequestID)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, resourceView)

	// invalid request id
	mockRequestID = "gd78tegd-0e737egd-jhdg"
	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(20111, requestStatusErrResponse))
	resourceView, err = client.GetRequestResourceView(mockRequestID)
	utils.AssertNotNilError(t, err)
	utils.AssertNil(t, resourceView)
}

func TestGetRequestResources(t *testing.T) {

	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	mockRequestID := "6ec160e5-41c5-4b1d-8ddc-e89c426957c6"
	path := fmt.Sprintf(GetRequestResourcesAPI, mockRequestID)
	url := client.BuildEncodedURL(path, nil)
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, resourcesResponse))

	resources, err := client.GetRequestResources(mockRequestID)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, resources)

	// invalid request id
	mockRequestID = "gd78tegd-0e737egd-jhdg"
	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(20111, requestStatusErrResponse))
	resources, err = client.GetRequestResources(mockRequestID)
	utils.AssertNotNilError(t, err)
	utils.AssertNil(t, resources)
}

func TestGetResourceActionTemplate(t *testing.T) {
	httpmock.ActivateNonDefault(client.Client)
	defer httpmock.DeactivateAndReset()

	mockResourceID := "0786a919-3545-408d-9091-cc8fe24e7790"
	mockActionID := "1a22752b-31a9-462e-a38a-e42b60c08a78"

	// test for delete action template
	getActionTemplatePath := fmt.Sprintf(GetActionTemplateAPI, mockResourceID, mockActionID)
	url := client.BuildEncodedURL(getActionTemplatePath, nil)
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, deleteActionTemplateResponse))

	actionTemplte, err := client.GetResourceActionTemplate(mockResourceID, mockActionID)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, actionTemplte)

	//test for reconfigure action tenplate
	httpmock.Reset()
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, reconfigureActionTemplateResponse))
	actionTemplte, err = client.GetResourceActionTemplate(mockResourceID, mockActionID)
	utils.AssertNilError(t, err)
	utils.AssertNotNil(t, actionTemplte)

	// invalid resource id
	mockResourceID = "gd78tegd-0e737egd-jhdg"
	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(10101, invalidResourceErrorResponse))
	actionTemplte, err = client.GetResourceActionTemplate(mockResourceID, mockActionID)
	utils.AssertNotNilError(t, err)
	utils.AssertNil(t, actionTemplte)

	//invalid action id
	mockActionID = "7364g-8736eg-87736"
	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(50505, systemExceptionResponse))
	actionTemplte, err = client.GetResourceActionTemplate(mockResourceID, mockActionID)
	utils.AssertNotNilError(t, err)
	utils.AssertNil(t, actionTemplte)

}
