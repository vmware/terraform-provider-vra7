package vra7

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	logging "github.com/op/go-logging"
	"github.com/terraform-providers/terraform-provider-vra7/sdk"
	"github.com/terraform-providers/terraform-provider-vra7/utils"
)

// error constants
const (
	ConfigInvalidError                = "The resource_configuration in the config file has invalid component name(s): %v "
	DestroyActionTemplateError        = "Error retrieving destroy action template for the deployment %v: %v "
	BusinessGroupIDNameNotMatchingErr = "The business group name %s and id %s does not belong to the same business group, provide either name or id"
	CatalogItemIDNameNotMatchingErr   = "The catalog item name %s and id %s does not belong to the same catalog item, provide either name or id"
)

var (
	log = logging.MustGetLogger(utils.LoggerID)
)

// ProviderSchema represents the information provided in the tf file
type ProviderSchema struct {
	CatalogItemName         string
	CatalogItemID           string
	Description             string
	Reasons                 string
	BusinessGroupID         string
	BusinessGroupName       string
	WaitTimeout             int
	RequestStatus           string
	DeploymentConfiguration map[string]interface{}
	DeploymentDestroy       bool
	Lease                   int
	DeploymentID            string
	ResourceConfiguration   []sdk.ResourceConfigurationStruct
}

func resourceVra7Deployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVra7DeploymentCreate,
		Read:   resourceVra7DeploymentRead,
		Update: resourceVra7DeploymentUpdate,
		Delete: resourceVra7DeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_item_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"catalog_item_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"reasons": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"businessgroup_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"businessgroup_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"deployment_configuration": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"deployment_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"resource_configuration": resourceConfigurationSchema(false),
			"lease_days": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"wait_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  15,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lease_start": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lease_end": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_status": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// Terraform call - terraform apply
// This function creates a new vRA 7 Deployment using configuration in a user's Terraform file.
// The Deployment is produced by invoking a catalog item that is specified in the configuration.
func resourceVra7DeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	log.Info("Creating the resource vra7_deployment...")
	vraClient := meta.(*sdk.APIClient)
	// Get client handle

	validityErr := checkConfigValuesValidity(d)
	if validityErr != nil {
		return validityErr
	}

	p, err := readProviderConfiguration(d, vraClient)
	if err != nil {
		return err
	}

	requestTemplate, validityErr := p.checkResourceConfigValidity(vraClient)
	if validityErr != nil {
		return validityErr
	}

	requestTemplate.Description = p.Description
	requestTemplate.Reasons = p.Reasons
	// if business group is not provided, the default business group in the request template is used
	if p.BusinessGroupID != "" {
		requestTemplate.BusinessGroupID = p.BusinessGroupID
	}
	requestTemplate.Data["_leaseDays"] = p.Lease
	for field, value := range p.DeploymentConfiguration {
		requestTemplate.Data[field] = utils.UnmarshalJSONStringIfNecessary(field, value)
	}

	for _, rConfig := range p.ResourceConfiguration {
		if rConfig.Cluster != 0 {
			rConfig.Configuration["_cluster"] = rConfig.Cluster
		}
		for propertyName, propertyValue := range rConfig.Configuration {
			requestTemplate.Data[rConfig.ComponentName] = updateRequestTemplate(
				requestTemplate.Data[rConfig.ComponentName].(map[string]interface{}),
				propertyName,
				propertyValue)
		}
	}

	log.Info("The updated catalog item request template  is %v\n", requestTemplate.Data)

	//Fire off a catalog item request to create a deployment.
	catalogRequest, err := vraClient.RequestCatalogItem(requestTemplate)

	if err != nil {
		return fmt.Errorf("The catalog item request failed with error %v", err)
	}
	_, err = waitForRequestCompletion(d, meta, catalogRequest.ID)
	if err != nil {
		return err
	}
	d.SetId(catalogRequest.ID)
	log.Info("Finished creating the resource vra7_deployment with request id %s", d.Id())
	return resourceVra7DeploymentRead(d, meta)
}

func updateRequestTemplate(templateInterface map[string]interface{}, field string, value interface{}) map[string]interface{} {
	replaced := utils.ReplaceValueInRequestTemplate(templateInterface, field, value)

	if !replaced {
		templateInterface["data"] = utils.AddValueToRequestTemplate(templateInterface["data"].(map[string]interface{}), field, value)
	}
	return templateInterface
}

// This function updates the state of a vRA 7 Deployment when changes to a Terraform file are applied.
//The update is performed on the Deployment using supported (day-2) actions.
func resourceVra7DeploymentUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Info("Updating the resource vra7_deployment with request id %s", d.Id())
	vraClient := meta.(*sdk.APIClient)
	// Get the ID of the catalog request that was used to provision this Deployment.
	catalogItemRequestID := d.Id()
	// Get client handle

	p, err := readProviderConfiguration(d, vraClient)
	if err != nil {
		return err
	}
	_, validityErr := p.checkResourceConfigValidity(vraClient)
	if validityErr != nil {
		return validityErr
	}

	//Change Lease Day 2 operation
	if d.HasChange("lease_days") {
		deploymentResourceActions, _ := vraClient.GetResourceActions(p.DeploymentID)
		deploymentActionsMap := GetActionNameIDMap(deploymentResourceActions)
		changeLeaseActionID := deploymentActionsMap["Change Lease"]
		if changeLeaseActionID != "" {
			resourceActionTemplate, _ := vraClient.GetResourceActionTemplate(p.DeploymentID, changeLeaseActionID)
			currTime, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))    // format 2020-04-06 00:15:44 -0700 PDT
			newLeaseEnd := currTime.Add(time.Hour * 24 * time.Duration(int64(p.Lease))) // format 2020-04-06 00:15:44 -0700 PDT
			extendLeaseTo := ConvertToISO8601(newLeaseEnd)
			log.Info("Starting Change Lease action on the deployment with id %v. The lease will be extended by %v days.", p.DeploymentID, p.Lease)
			_ = utils.ReplaceValueInRequestTemplate(
				resourceActionTemplate.Data, "provider-ExpirationDate", extendLeaseTo)
			requestID, err := vraClient.PostResourceAction(p.DeploymentID, changeLeaseActionID, resourceActionTemplate)
			if err != nil {
				log.Errorf("The change lease request failed with error: %v ", err)
				return err
			}
			_, err = waitForRequestCompletion(d, meta, requestID)
			if err != nil {
				log.Errorf("The change lease request failed with error: %v ", err)
				return err
			}
			log.Info("Successfully completed the Change Lease action for the deployment with id %v.", p.DeploymentID)
		}
	}

	// get the old and new resource_configuration data
	old, new := d.GetChange("resource_configuration")
	oldResourceConfigList := expandResourceConfiguration(old.(*schema.Set).List())
	newResourceConfigList := expandResourceConfiguration(new.(*schema.Set).List())

	if d.HasChange("resource_configuration") {
		for _, newResourceConfig := range newResourceConfigList {
			oldResourceConfig := GetResourceConfigurationByComponent(oldResourceConfigList, newResourceConfig.ComponentName)
			if oldResourceConfig.ComponentName != "" {
				deploymentResourceActions, _ := vraClient.GetResourceActions(p.DeploymentID)
				deploymentActionsMap := GetActionNameIDMap(deploymentResourceActions)
				if newResourceConfig.Cluster != 0 && oldResourceConfig.Cluster != newResourceConfig.Cluster {
					if oldResourceConfig.Cluster < newResourceConfig.Cluster && deploymentActionsMap[sdk.ScaleOut] != "" {
						// Scale Out Day 2 operation
						scaleOutActionID := deploymentActionsMap[sdk.ScaleOut]
						// get the action template for scale out
						resourceActionTemplate, err := vraClient.GetResourceActionTemplate(p.DeploymentID, scaleOutActionID)
						if err != nil {
							return err
						}
						// get the map from the action template corresponding to the key which is the component name
						actionTemplateDataMap := GetActionTemplateDataByComponent(resourceActionTemplate.Data, newResourceConfig.ComponentName)
						// update the template with the new cluster size
						log.Info("Starting Scale In action on the deployment with id %v for the component %v. The cluster size will be reduced from %v to %v.",
							p.DeploymentID, oldResourceConfig.ComponentName, oldResourceConfig.Cluster, newResourceConfig.Cluster)
						_ = utils.ReplaceValueInRequestTemplate(
							actionTemplateDataMap, "_cluster", newResourceConfig.Cluster)
						requestID, err := vraClient.PostResourceAction(p.DeploymentID, scaleOutActionID, resourceActionTemplate)
						if err != nil {
							log.Errorf("The scale out request failed with error: %v ", err)
							return err
						}
						log.Info("The Scale Out operation for the component %v has been submitted", newResourceConfig.ComponentName)
						_, err = waitForRequestCompletion(d, meta, requestID)
						if err != nil {
							log.Errorf("The scale out request failed with error: %v ", err)
							return err
						}
						log.Info("Successfully completed the Scale In action for the deployment with id %v.", p.DeploymentID)
					} else if oldResourceConfig.Cluster > newResourceConfig.Cluster && deploymentActionsMap[sdk.ScaleIn] != "" {
						// Scale In Day 2 operation
						scaleInActionID := deploymentActionsMap[sdk.ScaleIn]
						// get the action template for scale in
						resourceActionTemplate, err := vraClient.GetResourceActionTemplate(p.DeploymentID, scaleInActionID)
						if err != nil {
							return err
						}
						// get the map from the action template corresponding to the key which is the component name
						actionTemplateDataMap := GetActionTemplateDataByComponent(resourceActionTemplate.Data, newResourceConfig.ComponentName)
						// update the template with the new cluster size
						log.Info("Starting Scale Out action on the deployment with id %v for the component %v. The cluster size will be increased from %v to %v.",
							p.DeploymentID, oldResourceConfig.ComponentName, oldResourceConfig.Cluster, newResourceConfig.Cluster)
						_ = utils.ReplaceValueInRequestTemplate(
							actionTemplateDataMap, "_cluster", newResourceConfig.Cluster)
						requestID, err := vraClient.PostResourceAction(p.DeploymentID, scaleInActionID, resourceActionTemplate)
						if err != nil {
							log.Errorf("The scale in request failed with error: %v ", err)
							return err
						}
						log.Info("The Scale In operation for the component %v has been submitted", newResourceConfig.ComponentName)
						_, err = waitForRequestCompletion(d, meta, requestID)
						if err != nil {
							log.Errorf("The scale in request failed with error: %v ", err)
							return err
						}
						log.Info("Successfully completed the Scale Out action for the deployment with id %v.", p.DeploymentID)
					}
				}
			}
		}

		resources, err := vraClient.GetRequestResources(catalogItemRequestID)
		if err != nil {
			return fmt.Errorf("Error while getting resources for the request %v: %v  ", catalogItemRequestID, err.Error())
		}
		// Reconfigure Day 2 operation
		for _, resource := range resources.Content {
			oldResourceConfig := GetResourceByID(oldResourceConfigList, resource.ID)
			if oldResourceConfig.ComponentName != "" {
				vmResourceActions, _ := vraClient.GetResourceActions(oldResourceConfig.ResourceID)
				vmResourceActionsMap := GetActionNameIDMap(vmResourceActions)
				if vmResourceActionsMap[sdk.Reconfigure] != "" {
					reconfigureActionID := vmResourceActionsMap[sdk.Reconfigure]
					resourceActionTemplate, _ := vraClient.GetResourceActionTemplate(oldResourceConfig.ResourceID, reconfigureActionID)
					newResourceConfig := GetResourceConfigurationByComponent(newResourceConfigList, oldResourceConfig.ComponentName)
					configChanged := false
					actionTemplateDataMap := resourceActionTemplate.Data
					for propertyName, propertyValue := range newResourceConfig.Configuration {
						if oldResourceConfig.Configuration[propertyName] != propertyValue {
							_ = utils.ReplaceValueInRequestTemplate(
								actionTemplateDataMap, propertyName, propertyValue)
							if !configChanged {
								configChanged = true
							}
						}
					}
					if configChanged {
						log.Info("Starting Reconfigure action on the component %v.", oldResourceConfig.ComponentName)
						requestID, err := vraClient.PostResourceAction(oldResourceConfig.ResourceID, reconfigureActionID, resourceActionTemplate)
						if err != nil {
							log.Errorf("The reconfigure request failed with error: %v ", err)
							return err
						}
						log.Info("The Scale In operation for the component %v has been submitted", oldResourceConfig.ComponentName)
						_, err = waitForRequestCompletion(d, meta, requestID)
						if err != nil {
							log.Errorf("The reconfigure request for component %v failed with error: %v ", oldResourceConfig.ComponentName, err)
							return err
						}
						log.Info("Successfully completed the Reconfiguret action on the component %v.", oldResourceConfig.ComponentName)
					}
				}
			}
		}
	}
	log.Info("Finished updating the resource vra7_deployment with request id %s", d.Id())
	return resourceVra7DeploymentRead(d, meta)
}

// This function retrieves the latest state of a vRA 7 deployment. Terraform updates its state based on
// the information returned by this function.
func resourceVra7DeploymentRead(d *schema.ResourceData, meta interface{}) error {
	log.Info("Reading the resource vra7_deployment with request id %s ", d.Id())
	vraClient := meta.(*sdk.APIClient)

	// Get the ID of the catalog request that was used to provision this Deployment. This id
	// will remain the same for this deployment across any actions on the machines like reconfigure, etc.
	catalogItemRequestID := d.Id()

	requestResourceView, errTemplate := vraClient.GetRequestResourceView(catalogItemRequestID)
	if requestResourceView != nil && len(requestResourceView.Content) == 0 {
		//If resource does not exists then unset the resource ID from state file
		d.SetId("")
		return fmt.Errorf("The resource cannot be found")
	}
	if errTemplate != nil || len(requestResourceView.Content) == 0 {
		return fmt.Errorf("Resource view failed to load with the error %v", errTemplate)
	}

	clusterCountMap := make(map[string]int)
	var resourceConfigList []sdk.ResourceConfigurationStruct
	for _, resource := range requestResourceView.Content {
		rMap := resource.(map[string]interface{})
		resourceType := rMap["resourceType"].(string)
		name := rMap["name"].(string)
		dateCreated := rMap["dateCreated"].(string)
		lastUpdated := rMap["lastUpdated"].(string)
		resourceID := rMap["resourceId"].(string)
		requestID := rMap["requestId"].(string)
		requestState := rMap["requestState"].(string)

		// if the resource type is VMs, update the resource_configuration attribute
		if resourceType == sdk.InfrastructureVirtual {
			data := rMap["data"].(map[string]interface{})
			componentName := data["Component"].(string)
			parentResourceID := rMap["parentResourceId"].(string)
			var resourceConfigStruct sdk.ResourceConfigurationStruct
			resourceConfigStruct.Configuration = data
			resourceConfigStruct.ComponentName = componentName
			resourceConfigStruct.Name = name
			resourceConfigStruct.DateCreated = dateCreated
			resourceConfigStruct.LastUpdated = lastUpdated
			resourceConfigStruct.ResourceID = resourceID
			resourceConfigStruct.ResourceType = resourceType
			resourceConfigStruct.RequestID = requestID
			resourceConfigStruct.RequestState = requestState
			resourceConfigStruct.ParentResourceID = parentResourceID
			resourceConfigStruct.IPAddress = data["ip_address"].(string)

			if rMap["description"] != nil {
				resourceConfigStruct.Description = rMap["description"].(string)
			}
			if rMap["status"] != nil {
				resourceConfigStruct.Status = rMap["status"].(string)
			}
			// the cluster value is calculated from the map based on the component name as the
			// resourceViews API does not return that information
			clusterCountMap[componentName] = clusterCountMap[componentName] + 1
			resourceConfigList = append(resourceConfigList, resourceConfigStruct)

		} else if resourceType == sdk.DeploymentResourceType {

			leaseMap := rMap["lease"].(map[string]interface{})
			leaseStart := leaseMap["start"].(string)
			d.Set("lease_start", leaseStart)
			// if the lease never expires, the end date will be null
			if leaseMap["end"] != nil {
				leaseEnd := leaseMap["end"].(string)
				d.Set("lease_end", leaseEnd)
				// the lease_days are calculated from the current time and lease_end dates as the resourceViews API does not return that information
				currTime, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				endTime, _ := time.Parse(time.RFC3339, leaseEnd)
				diff := endTime.Sub(currTime)
				d.Set("lease_days", int(diff.Hours()/24))
				// end
			} else {
				d.Set("lease_days", nil) // set lease days to nil if lease_end is nil
			}

			d.Set("catalog_item_id", rMap["catalogItemId"].(string))
			d.Set("catalog_item_name", rMap["catalogItemLabel"].(string))
			d.Set("deployment_id", resourceID)
			d.Set("date_created", dateCreated)
			d.Set("last_updated", lastUpdated)
			d.Set("tenant_id", rMap["tenantId"].(string))
			d.Set("owners", rMap["owners"].([]interface{}))
			d.Set("name", name)
			d.Set("businessgroup_id", rMap["businessGroupId"].(string))
			if rMap["description"] != nil {
				d.Set("description", rMap["description"].(string))
			}
			if rMap["status"] != nil {
				d.Set("request_status", rMap["status"].(string))
			}
		}
	}
	if err := d.Set("resource_configuration", flattenResourceConfigurations(resourceConfigList, clusterCountMap)); err != nil {
		return fmt.Errorf("error setting resource configuration - error: %v", err)
	}

	log.Info("Finished reading the resource vra7_deployment with request id %s", d.Id())
	return nil
}

//Function use - To delete resources which are created by terraform and present in state file
func resourceVra7DeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	log.Info("Deleting the resource vra7_deployment with request id %s", d.Id())
	vraClient := meta.(*sdk.APIClient)

	// Throw an error if request ID has no value or empty value
	if len(d.Id()) == 0 {
		return fmt.Errorf("Resource not found")
	}

	p, err := readProviderConfiguration(d, vraClient)
	if err != nil {
		return err
	}

	deploymentID := d.Get("deployment_id").(string)
	deploymentResourceActions, _ := vraClient.GetResourceActions(deploymentID)
	deploymentActionsMap := GetActionNameIDMap(deploymentResourceActions)
	destroyActionID := deploymentActionsMap["Destroy"]
	if p.DeploymentDestroy && destroyActionID != "" {
		resourceActionTemplate, _ := vraClient.GetResourceActionTemplate(deploymentID, destroyActionID)
		requestID, err := vraClient.PostResourceAction(deploymentID, destroyActionID, resourceActionTemplate)
		if err != nil {
			log.Errorf("The destroy request failed with error: %v ", err)
			return err
		}
		status, err := waitForRequestCompletion(d, meta, requestID)
		if err != nil {
			log.Errorf("The destroy request failed with error: %v ", err)
			return err
		}
		if status == sdk.Successful {
			d.SetId("")
		}
	}
	log.Info("Finished destroying the resource vra7_deployment with request id %s", d.Id())
	return nil
}

// check if the resource configuration is valid in the terraform config file
func (p *ProviderSchema) checkResourceConfigValidity(client *sdk.APIClient) (*sdk.CatalogItemRequestTemplate, error) {
	log.Info("Checking if the terraform config file is valid")

	// Get request template for catalog item.
	requestTemplate, err := client.GetCatalogItemRequestTemplate(p.CatalogItemID)
	if err != nil {
		return nil, err
	}
	log.Info("The request template data corresponding to the catalog item %v is: \n %v\n", p.CatalogItemID, requestTemplate.Data)

	// Get all component names in the blueprint corresponding to the catalog item.
	componentSet := make(map[string]bool)
	for field := range requestTemplate.Data {
		if reflect.ValueOf(requestTemplate.Data[field]).Kind() == reflect.Map {
			componentSet[field] = true
		}
	}
	log.Info("The component name(s) in the blueprint corresponding to the catalog item: %v\n", componentSet)

	var invalidKeys []string
	// check if the component of resource_configuration map exists in the componentSet
	// retrieved from catalog item request template
	for _, k := range p.ResourceConfiguration {
		if _, ok := componentSet[k.ComponentName]; !ok {
			invalidKeys = append(invalidKeys, k.ComponentName)
		}
	}
	// there are invalid resource config keys in the terraform config file, abort and throw an error
	if len(invalidKeys) > 0 {
		return nil, fmt.Errorf(ConfigInvalidError, strings.Join(invalidKeys, ", "))
	}

	return requestTemplate, nil
}

// check if the values provided in the config file are valid and set
// them in the resource schema. Requires to call APIs
func checkConfigValuesValidity(d *schema.ResourceData) error {

	catalogItemName := d.Get("catalog_item_name").(string)
	catalogItemID := d.Get("catalog_item_id").(string)
	businessgroupName := d.Get("businessgroup_name").(string)
	businessgroupID := d.Get("businessgroup_id").(string)

	// If catalogItemID and catalogItemName both not provided then return an error
	if catalogItemID == "" && catalogItemName == "" {
		return fmt.Errorf("Provide either a catalog_item_name or a catalog_item_id in the configuration")
	}

	// If both catalog_item_name and catalogItemName return an error
	if catalogItemID != "" && catalogItemName != "" {
		return fmt.Errorf("Provide either a catalog_item_name or a catalog_item_id in the configuration")
	}

	// If both businessgroupID and businessgroupName return an error
	if businessgroupID != "" && businessgroupName != "" {
		return fmt.Errorf("Provide either a businessgroup_id or a businessgroup_name in the configuration")
	}
	return nil
}

// read the config file
func readProviderConfiguration(d *schema.ResourceData, vraClient *sdk.APIClient) (*ProviderSchema, error) {

	log.Info("Reading the provider configuration data.....")
	providerSchema := ProviderSchema{
		CatalogItemName:         strings.TrimSpace(d.Get("catalog_item_name").(string)),
		CatalogItemID:           strings.TrimSpace(d.Get("catalog_item_id").(string)),
		Description:             strings.TrimSpace(d.Get("description").(string)),
		Reasons:                 strings.TrimSpace(d.Get("reasons").(string)),
		BusinessGroupName:       strings.TrimSpace(d.Get("businessgroup_name").(string)),
		BusinessGroupID:         strings.TrimSpace(d.Get("businessgroup_id").(string)),
		Lease:                   d.Get("lease_days").(int),
		DeploymentID:            strings.TrimSpace(d.Get("deployment_id").(string)),
		WaitTimeout:             d.Get("wait_timeout").(int) * 60,
		ResourceConfiguration:   expandResourceConfiguration(d.Get("resource_configuration").(*schema.Set).List()),
		DeploymentDestroy:       d.Get("deployment_destroy").(bool),
		DeploymentConfiguration: d.Get("deployment_configuration").(map[string]interface{}),
	}

	// if catalog item name is provided, fetch the catalog item id
	if len(providerSchema.CatalogItemName) > 0 {
		id, err := vraClient.ReadCatalogItemByName(providerSchema.CatalogItemName)
		if err != nil {
			return &providerSchema, err
		}
		providerSchema.CatalogItemID = id
	}

	// get the business group id from name
	if len(providerSchema.BusinessGroupName) > 0 {
		id, err := vraClient.GetBusinessGroupID(providerSchema.BusinessGroupName, vraClient.Tenant)
		if err != nil {
			return &providerSchema, err
		}
		providerSchema.BusinessGroupID = id
	}

	log.Info("The values provided in the TF config file is: \n %v ", providerSchema)
	return &providerSchema, nil
}

// check the request status on apply update and destroy
func waitForRequestCompletion(d *schema.ResourceData, meta interface{}, requestID string) (string, error) {
	vraClient := meta.(*sdk.APIClient)
	waitTimeout := d.Get("wait_timeout").(int) * 60
	sleepFor := 30
	status := ""
	for i := 0; i < waitTimeout/sleepFor; i++ {
		log.Info("Waiting for %d seconds before checking request status.", sleepFor)
		time.Sleep(time.Duration(sleepFor) * time.Second)
		reqestStatusView, _ := vraClient.GetRequestStatus(requestID)
		status = reqestStatusView.Phase
		d.Set("request_status", status)
		log.Info("Checking to see the status of the request. Status: %s.", status)
		if status == sdk.Successful {
			log.Info("Request is SUCCESSFUL.")
			return sdk.Successful, nil
		} else if status == sdk.Failed {
			return sdk.Failed, fmt.Errorf("Request failed \n %v ", reqestStatusView.RequestCompletion.CompletionDetails)
		} else if status == sdk.InProgress {
			log.Info("The request is still IN PROGRESS.")
		} else {
			log.Info("Request status: %s.", status)
		}
	}
	// The execution has timed out while still IN PROGRESS.
	// The user will need to use 'terraform refresh' at a later point to resolve this.
	return "", fmt.Errorf("Request has timed out with status %s. \nRun terraform refresh to get the latest state of your request", status)
}

// GetActionTemplateDataByComponent return the map corresponding the component name in the template data
func GetActionTemplateDataByComponent(actionTemplate map[string]interface{}, componentName string) map[string]interface{} {
	actionTemplateDataByComponent := make(map[string]interface{})
	for key, value := range actionTemplate {
		if key == componentName && reflect.ValueOf(value).Kind() == reflect.Map {
			actionTemplateDataByComponent = value.(map[string]interface{})
			break
		}
	}
	return actionTemplateDataByComponent
}

// GetResourceConfigurationByComponent returns the resource_configuration corresponding the component
func GetResourceConfigurationByComponent(resourceConfigurationList []sdk.ResourceConfigurationStruct, component string) sdk.ResourceConfigurationStruct {
	var resourceConfig sdk.ResourceConfigurationStruct
	for _, rConfig := range resourceConfigurationList {
		if rConfig.ComponentName == component {
			resourceConfig = rConfig
		}
	}
	return resourceConfig
}

// GetActionNameIDMap returns a map of Action name and id
func GetActionNameIDMap(resourceActions []sdk.Operation) map[string]string {
	actionNameIDMap := make(map[string]string)
	for _, action := range resourceActions {
		actionNameIDMap[action.Name] = action.ID
	}
	return actionNameIDMap
}

// GetResourceByID return the resoirce config struct object filtered by ID
func GetResourceByID(resourceConfigStructList []sdk.ResourceConfigurationStruct, resourceID string) sdk.ResourceConfigurationStruct {
	var resourceConfigStruct sdk.ResourceConfigurationStruct
	for _, resourceStruct := range resourceConfigStructList {
		if resourceStruct.ResourceID == resourceID {
			resourceConfigStruct = resourceStruct
		}
	}
	return resourceConfigStruct
}

// ConvertToISO8601 converts time to the format ISO8601 (2020-04-16T00:15:44.700Z) that vRA 7.x understands
func ConvertToISO8601(givenTime time.Time) string {
	rfc339Time, _ := time.Parse(time.RFC3339, givenTime.Format(time.RFC3339)) // 2020-04-06 00:15:44 -0700 PDT
	testArray := strings.Fields(rfc339Time.String())                          // ["2020-04-06", "00:15:44", "-0700", "PDT"]
	length := len(testArray[2])                                               // length of -0700
	last3 := testArray[2][length-3 : length]                                  // 700
	iso8601Time := testArray[0] + "T" + testArray[1] + "." + last3 + "Z"      // 2020-04-16T00:15:44.700Z
	return iso8601Time
}
