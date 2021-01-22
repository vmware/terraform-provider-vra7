package sdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/vmware/terraform-provider-vra7/utils"
)

// API constants
const (
	IdentityAPI                 = "/identity/api"
	Tokens                      = IdentityAPI + "/tokens"
	Tenants                     = IdentityAPI + "/tenants"
	CatalogService              = "/catalog-service"
	CatalogServiceAPI           = CatalogService + "/api"
	Consumer                    = CatalogServiceAPI + "/consumer"
	ConsumerRequests            = Consumer + "/requests"
	ConsumerResources           = Consumer + "/resources"
	EntitledCatalogItems        = Consumer + "/entitledCatalogItems"
	EntitledCatalogItemViewsAPI = Consumer + "/entitledCatalogItemViews"
	GetResourceAPI              = Consumer + "/resources/%s"
	GetRequestResourcesAPI      = ConsumerRequests + "/" + "%s" + "/resources"
	ResourceActions             = ConsumerResources + "/" + "%s" + "/actions"
	PostActionTemplateAPI       = ResourceActions + "/" + "%s" + "/requests"
	GetActionTemplateAPI        = PostActionTemplateAPI + "/template"
	GetRequestResourceViewAPI   = ConsumerRequests + "/" + "%s" + "/resourceViews"
	RequestTemplateAPI          = EntitledCatalogItems + "/" + "%s" + "/requests/template"
	GetDeploymentAPI            = Consumer + "/deployments/%s"

	InProgress             = "IN_PROGRESS"
	Successful             = "SUCCESSFUL"
	Failed                 = "FAILED"
	Submitted              = "SUBMITTED"
	InfrastructureVirtual  = "Infrastructure.Virtual"
	DeploymentResourceType = "composition.resource.type.deployment"
	Component              = "Component"
	Reconfigure            = "Reconfigure"
	Destroy                = "Destroy"
	ScaleOut               = "Scale Out"
	ScaleIn                = "Scale In"
	DeploymentDestroy      = "Deployment Destroy"
)

// GetCatalogItemRequestTemplate - Call to retrieve a request template for a catalog item.
func (c *APIClient) GetCatalogItemRequestTemplate(catalogItemID string) (*CatalogItemRequestTemplate, error) {

	// Form a path to read catalog request template via REST call
	path := fmt.Sprintf(RequestTemplateAPI, catalogItemID)
	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var requestTemplate CatalogItemRequestTemplate
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &requestTemplate)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &requestTemplate, nil
}

// ReadCatalogItemNameByID - This function returns the catalog item name using catalog item ID
func (c *APIClient) ReadCatalogItemNameByID(catalogItemID string) (string, error) {

	path := fmt.Sprintf(EntitledCatalogItems+"/"+"%s", catalogItemID)
	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return "", respErr
	}

	var response CatalogItem
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return "", unmarshallErr
	}
	return response.CatalogItem.Name, nil
}

// ReadCatalogItemByName to read id of catalog from vRA using catalog_name
func (c *APIClient) ReadCatalogItemByName(catalogName string) (string, error) {

	// reading the first page to get the total number of pages
	entitledCatalogItems, err := c.readCatalogItemsByPage(1)
	if err != nil {
		return "", err
	}

	for page := 1; page <= entitledCatalogItems.Metadata.TotalPages; page++ {
		entitledCatalogItemViews, err := c.readCatalogItemsByPage(page)
		if err != nil {
			return "", err
		}
		catalogItemsArray := entitledCatalogItemViews.Content.([]interface{})
		for i := range catalogItemsArray {
			catalogItem := catalogItemsArray[i].(map[string]interface{})
			name := catalogItem["name"].(string)
			if name == catalogName {
				return catalogItem["catalogItemId"].(string), nil
			}
		}

	}
	return "", fmt.Errorf("Catalog item, %s not found", catalogName)
}

// ReadCatalogItemsByPage return catalogItems by page
func (c *APIClient) readCatalogItemsByPage(i int) (*EntitledCatalogItemViews, error) {
	url := c.BuildEncodedURL(EntitledCatalogItemViewsAPI, map[string]string{
		"page": strconv.Itoa(i)})
	resp, respErr := c.Get(url, nil)
	if respErr != nil || resp.StatusCode != 200 {
		return nil, respErr
	}

	var template EntitledCatalogItemViews
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &template)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}

	return &template, nil
}

// GetBusinessGroupID retrieves business group id from business group name
func (c *APIClient) GetBusinessGroupID(businessGroupName string, tenant string) (string, error) {

	path := Tenants + "/" + tenant + "/subtenants"

	log.Info("Fetching business group id from name..GET %s ", path)

	url := c.BuildEncodedURL(path, nil)

	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return "", respErr
	}

	var businessGroups BusinessGroups
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &businessGroups)
	if unmarshallErr != nil {
		return "", unmarshallErr
	}

	if len(businessGroups.Content) == 0 {
		path = path + "/membership"

		membershipURL := c.BuildEncodedURL(path, nil)

		resp, respErr := c.Get(membershipURL, nil)
		if respErr != nil {
			return "", respErr
		}
		unmarshallErr := utils.UnmarshalJSON(resp.Body, &businessGroups)
		if unmarshallErr != nil {
			return "", unmarshallErr
		}
	}

	// BusinessGroups array will contain only one BusinessGroup element containing the BG
	// with the name businessGroupName.
	// Fetch the id of that
	for _, businessGroup := range businessGroups.Content {
		if businessGroup.Name == businessGroupName {
			log.Info("Found the business group id of the group %s: %s ", businessGroupName, businessGroup.ID)
			return businessGroup.ID, nil
		}
	}
	return "", fmt.Errorf("No business group found with name: %s ", businessGroupName)
}

// GetRequestStatus - To read request status of resource
// which is used to show information to user post create call.
func (c *APIClient) GetRequestStatus(requestID string) (*RequestStatusView, error) {
	//Form a URL to read request status
	path := fmt.Sprintf(ConsumerRequests+"/"+"%s", requestID)
	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var response RequestStatusView
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

// GetRequestResourceView retrieves the resources that were provisioned as a result of a given request.
func (c *APIClient) GetRequestResourceView(catalogRequestID string, pageID int) (*RequestResourceView, error) {
	path := fmt.Sprintf(GetRequestResourceViewAPI, catalogRequestID)
	queryParam := make(map[string]string)
	queryParam["page"] = strconv.Itoa(pageID)
	url := c.BuildEncodedURL(path, queryParam)

	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}
	var response RequestResourceView
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

// RequestCatalogItem - Make a catalog request.
func (c *APIClient) RequestCatalogItem(requestTemplate *CatalogItemRequestTemplate) (*CatalogRequest, error) {
	//Form a path to set a REST call to create a machine
	path := fmt.Sprintf(EntitledCatalogItems+"/"+"%s"+
		"/requests", requestTemplate.CatalogItemID)

	buffer, _ := utils.MarshalToJSON(requestTemplate)
	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Post(url, buffer, nil)
	if respErr != nil {
		return nil, respErr
	}

	var response CatalogRequest
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

// GetRequestResources get the resource actions allowed for a resource
func (c *APIClient) GetRequestResources(catalogItemRequestID string) (*Resources, error) {
	path := fmt.Sprintf(GetRequestResourcesAPI, catalogItemRequestID)

	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var requestResources Resources
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &requestResources)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &requestResources, nil
}

// GetResource get the resource actions allowed for a resource
func (c *APIClient) GetResource(resourceID string) (*ResourceContent, error) {
	path := fmt.Sprintf(GetResourceAPI, resourceID)

	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var resource ResourceContent
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &resource)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &resource, nil
}

// GetResourceActions get the resource actions allowed for a resource
func (c *APIClient) GetResourceActions(resourceID string) ([]Operation, error) {
	path := fmt.Sprintf(ResourceActions, resourceID)

	url := c.BuildEncodedURL(path, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var reqResponse RequestResponse
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &reqResponse)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	var resourceActions []Operation
	for _, item := range reqResponse.Content {
		action := &Operation{}
		err := mapstructure.Decode(item, &action)
		if err != nil {
			return nil, err
		}
		resourceActions = append(resourceActions, *action)
	}
	return resourceActions, nil
}

// GetResourceActionTemplate get the action template corresponding to the action id
func (c *APIClient) GetResourceActionTemplate(resourceID, actionID string) (*ResourceActionTemplate, error) {
	getActionTemplatePath := fmt.Sprintf(GetActionTemplateAPI, resourceID, actionID)
	log.Info("Call GET to fetch the action template %v ", getActionTemplatePath)
	url := c.BuildEncodedURL(getActionTemplatePath, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var resourceActionTemplate ResourceActionTemplate
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &resourceActionTemplate)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &resourceActionTemplate, nil
}

// PostResourceAction updates the resource
func (c *APIClient) PostResourceAction(resourceID, actionID string, resourceActionTemplate *ResourceActionTemplate) (string, error) {

	postActionTemplatePath := fmt.Sprintf(PostActionTemplateAPI, resourceID, actionID)
	buffer, _ := utils.MarshalToJSON(resourceActionTemplate)
	url := c.BuildEncodedURL(postActionTemplatePath, nil)
	resp, respErr := c.Post(url, buffer, nil)
	if respErr != nil || resp.StatusCode != 201 {
		return "", respErr
	}

	requestURL := resp.Location
	i := strings.LastIndex(requestURL, "/")
	requestID := requestURL[i+1:]

	return requestID, nil
}

// GetDeploymentIDFromRequest - get depoyment id from request id
func (c *APIClient) GetDeploymentIDFromRequest(requestID string) (string, error) {
	currentPage := 1
	totalPages := 1
	deploymentID := ""

	for currentPage <= totalPages {
		requestResourceView, err := c.GetRequestResourceView(requestID, currentPage)

		if requestResourceView != nil && len(requestResourceView.Content) == 0 {
			return "", fmt.Errorf("The resource cannot be found")
		}

		if err != nil || len(requestResourceView.Content) == 0 {
			return "", fmt.Errorf("Resource view failed to load with the error %v", err)
		}

		currentPage = requestResourceView.MetaData.Number + 1
		totalPages = requestResourceView.MetaData.TotalPages

		for _, resource := range requestResourceView.Content {
			// map containing the content of a resourceView response
			rMap := resource.(map[string]interface{})
			resourceType := rMap["resourceType"].(string)
			if resourceType == DeploymentResourceType {
				deploymentID = rMap["resourceId"].(string)
				break
			}
		}
	}
	return deploymentID, nil
}

// GetDeployment get the deployment with deploymentId
func (c *APIClient) GetDeployment(deploymentID string) (*Deployment, error) {
	deploymentAPIPath := fmt.Sprintf(GetDeploymentAPI, deploymentID)
	log.Info("Call GET to fetch the deployment %v ", deploymentAPIPath)
	url := c.BuildEncodedURL(deploymentAPIPath, nil)
	resp, respErr := c.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var deployment Deployment
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &deployment)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &deployment, nil
}
