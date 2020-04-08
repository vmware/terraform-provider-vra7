package sdk

import (
	"time"
)

// ResourceConfigurationStruct - structure representing the resource_configuration
type ResourceConfigurationStruct struct {
	ComponentName    string                 `json:"component_name,omitempty"`
	Cluster          int                    `json:"cluster,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Name             string                 `json:"name,omitempty"`
	ResourceID       string                 `json:"resource_id,omitempty"`
	Status           string                 `json:"status,omitempty"`
	RequestID        string                 `json:"request_id,omitempty"`
	RequestState     string                 `json:"request_state,omitempty"`
	ResourceType     string                 `json:"resource_type,omitempty"`
	Configuration    map[string]interface{} `json:"configuration,omitempty"`
	DateCreated      string                 `json:"last_created,omitempty"`
	LastUpdated      string                 `json:"last_updated,omitempty"`
	ParentResourceID string                 `json:"parent_resource_id,omitempty"`
	IPAddress        string                 `json:"ip_address,omitempty"`
}

// RequestResponse is the response structure of any request
type RequestResponse struct {
	Content []interface{} `json:"content,omitempty"`
	Links   []interface{} `json:"links,omitempty"`
}

// ResourceActionTemplate - is used to store information
// related to resource action template information.
type ResourceActionTemplate struct {
	Type        string                 `json:"type,omitempty"`
	ResourceID  string                 `json:"resourceId,omitempty"`
	ActionID    string                 `json:"actionId,omitempty"`
	Description string                 `json:"description,omitempty"`
	Reasons     string                 `json:"reasons,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// RequestStatusView - used to store REST response of
// request triggered against any resource.
type RequestStatusView struct {
	RequestCompletion struct {
		RequestCompletionState string `json:"requestCompletionState"`
		CompletionDetails      string `json:"CompletionDetails"`
	} `json:"requestCompletion"`
	Phase string `json:"phase"`
}

// BusinessGroups - list of business groups
type BusinessGroups struct {
	Content []BusinessGroup `json:"content,omitempty"`
}

// BusinessGroup - detail view of a business group
type BusinessGroup struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// RequestResourceView - resource view of a provisioned request
type RequestResourceView struct {
	Content []interface{} `json:"content,omitempty"`
	Links   []interface{} `json:"links,omitempty"`
}

// Resources - Retrieves the resources that were provisioned as a result of a given request.
// Also returns the actions allowed on the resources and their templates
type Resources struct {
	Links   []interface{}     `json:"links,omitempty"`
	Content []ResourceContent `json:"content,omitempty"`
}

// ResourceContent - Detailed view of the resource provisioned and the operation allowed
type ResourceContent struct {
	ID              string          `json:"id,omitempty"`
	Name            string          `json:"name,omitempty"`
	ResourceTypeRef ResourceTypeRef `json:"resourceTypeRef,omitempty"`
	Status          string          `json:"status,omitempty"`
	RequestID       string          `json:"requestId,omitempty"`
	RequestState    string          `json:"requestState,omitempty"`
	Operations      []Operation     `json:"operations,omitempty"`
	ResourceData    ResourceDataMap `json:"resourceData,omitempty"`
}

// ResourceTypeRef - type of resource (deployment, or machine, etc)
type ResourceTypeRef struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

// Operation - detailed view of an operation allowed on a resource
type Operation struct {
	Name        string `json:"name,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// ResourceDataMap - properties of a provisioned resource
type ResourceDataMap struct {
	Entries []ResourceDataEntry `json:"entries,omitempty"`
}

// ResourceDataEntry - the property key and value of a resource
type ResourceDataEntry struct {
	Key   string                 `json:"key,omitempty"`
	Value map[string]interface{} `json:"value,omitempty"`
}

//CatalogRequest - A structure that captures a vRA catalog request.
type CatalogRequest struct {
	ID           string      `json:"id"`
	IconID       string      `json:"iconId"`
	Version      int         `json:"version"`
	State        string      `json:"state"`
	Description  string      `json:"description"`
	Reasons      interface{} `json:"reasons"`
	RequestedFor string      `json:"requestedFor"`
	RequestedBy  string      `json:"requestedBy"`
	Organization struct {
		TenantRef      string `json:"tenantRef"`
		TenantLabel    string `json:"tenantLabel"`
		SubtenantRef   string `json:"subtenantRef"`
		SubtenantLabel string `json:"subtenantLabel"`
	} `json:"organization"`

	RequestorEntitlementID   string                 `json:"requestorEntitlementId"`
	PreApprovalID            string                 `json:"preApprovalId"`
	PostApprovalID           string                 `json:"postApprovalId"`
	DateCreated              time.Time              `json:"dateCreated"`
	LastUpdated              time.Time              `json:"lastUpdated"`
	DateSubmitted            time.Time              `json:"dateSubmitted"`
	DateApproved             time.Time              `json:"dateApproved"`
	DateCompleted            time.Time              `json:"dateCompleted"`
	Quote                    interface{}            `json:"quote"`
	RequestData              map[string]interface{} `json:"requestData"`
	RequestCompletion        string                 `json:"requestCompletion"`
	RetriesRemaining         int                    `json:"retriesRemaining"`
	RequestedItemName        string                 `json:"requestedItemName"`
	RequestedItemDescription string                 `json:"requestedItemDescription"`
	Components               string                 `json:"components"`
	StateName                string                 `json:"stateName"`

	CatalogItemProviderBinding struct {
		BindingID   string `json:"bindingId"`
		ProviderRef struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"providerRef"`
	} `json:"catalogItemProviderBinding"`

	Phase           string `json:"phase"`
	ApprovalStatus  string `json:"approvalStatus"`
	ExecutionStatus string `json:"executionStatus"`
	WaitingStatus   string `json:"waitingStatus"`
	CatalogItemRef  struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	} `json:"catalogItemRef"`
}

//CatalogItemRequestTemplate - A structure that captures a catalog request template, to be filled in and POSTED.
type CatalogItemRequestTemplate struct {
	Type            string                 `json:"type,omitempty"`
	CatalogItemID   string                 `json:"catalogItemId,omitempty"`
	RequestedFor    string                 `json:"requestedFor,omitempty"`
	BusinessGroupID string                 `json:"businessGroupId,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Reasons         string                 `json:"reasons,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

//catalogName - This struct holds catalog name from json response.
type catalogName struct {
	Name string `json:"name"`
	ID   string `json:"catalogItemId"`
}

//CatalogItem - This struct holds the value of response of catalog item list
type CatalogItem struct {
	CatalogItem catalogName `json:"catalogItem"`
}

// EntitledCatalogItemViews represents catalog items in an active state, the current user
// is entitled to consume
type EntitledCatalogItemViews struct {
	Links    interface{} `json:"links"`
	Content  interface{} `json:"content"`
	Metadata Metadata    `json:"metadata"`
}

// Metadata - Metadata  used to store metadata of resource list response
type Metadata struct {
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Number        int `json:"number"`
}
