package vra7

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
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
	log       = logging.MustGetLogger(utils.LoggerID)
	vraClient *sdk.APIClient
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
	FailedMessage           string
	DeploymentConfiguration map[string]interface{}
	ResourceConfiguration   map[string]interface{}
}

func resourceVra7Deployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVra7DeploymentCreate,
		Read:   resourceVra7DeploymentRead,
		Update: resourceVra7DeploymentUpdate,
		Delete: resourceVra7DeploymentDelete,

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
			"wait_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  15,
			},
			"request_status": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"failed_message": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
				Optional: true,
			},
			"deployment_configuration": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"resource_configuration": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

// byLength type to sort component name list by it's name length
type byLength []string

func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}
func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Terraform call - terraform apply
// This function creates a new vRA 7 Deployment using configuration in a user's Terraform file.
// The Deployment is produced by invoking a catalog item that is specified in the configuration.
func resourceVra7DeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	vraClient = meta.(*sdk.APIClient)
	// Get client handle
	p := readProviderConfiguration(d)

	requestTemplate, validityErr := p.checkConfigValuesValidity(d)
	if validityErr != nil {
		return validityErr
	}
	validityErr = p.checkResourceConfigValidity(requestTemplate)
	if validityErr != nil {
		return validityErr
	}

	requestTemplate.Description = p.Description
	requestTemplate.Reasons = p.Reasons

	updateRequestTemplateFromDeploymentConfiguration(p.DeploymentConfiguration, requestTemplate)

	// Get all component names in the blueprint corresponding to the catalog item.
	var componentNameList []string
	for field := range requestTemplate.Data {
		if reflect.ValueOf(requestTemplate.Data[field]).Kind() == reflect.Map {
			componentNameList = append(componentNameList, field)
		}
	}
	log.Info("createResource->key_list %v\n", componentNameList)

	// Arrange component names in descending order of text length.
	// Component names are sorted this way because '.', which is used as a separator, may also occur within
	// component names. In these situations, the longest name match that includes '.'s should win.
	sort.Sort(byLength(componentNameList))

	//Update request template field values with values from user configuration.
	for configKey, configValue := range p.ResourceConfiguration {
		for _, componentName := range componentNameList {
			// User-supplied resource configuration keys are expected to be of the form:
			//     <component name>.<property name>.
			// Extract the property names and values for each component in the blueprint, and add/update
			// them in the right location in the request template.
			if strings.HasPrefix(configKey, componentName) {
				propertyName := strings.TrimPrefix(configKey, componentName+".")
				if len(propertyName) == 0 {
					return fmt.Errorf(
						"resource_configuration key is not in correct format. Expected %s to start with %s",
						configKey, componentName+".")
				}
				// Function call which changes request template field values with user-supplied values
				requestTemplate.Data[componentName] = updateRequestTemplate(
					requestTemplate.Data[componentName].(map[string]interface{}),
					propertyName,
					configValue)
				break
			}
		}
	}

	log.Info("Updated template - %v\n", requestTemplate.Data)

	//Fire off a catalog item request to create a deployment.
	catalogRequest, err := vraClient.RequestCatalogItem(requestTemplate)

	if err != nil {
		log.Errorf("Resource Machine Request Failed: %v", err)
		return fmt.Errorf("Resource Machine Request Failed: %v", err)
	}
	_, err = waitForRequestCompletion(d, meta, catalogRequest.ID)
	if err != nil {
		return err
	}
	d.SetId(catalogRequest.ID)
	return resourceVra7DeploymentRead(d, meta)
}

func updateRequestTemplate(templateInterface map[string]interface{}, field string, value interface{}) map[string]interface{} {
	var replaced bool
	templateInterface, replaced = utils.ReplaceValueInRequestTemplate(templateInterface, field, value)

	if !replaced {
		templateInterface["data"] = utils.AddValueToRequestTemplate(templateInterface["data"].(map[string]interface{}), field, value)
	}
	return templateInterface
}

// Terraform call - terraform apply
// This function updates the state of a vRA 7 Deployment when changes to a Terraform file are applied.
// The update is performed on the Deployment using supported (day-2) actions.
func resourceVra7DeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	vraClient = meta.(*sdk.APIClient)
	// Get the ID of the catalog request that was used to provision this Deployment.
	catalogItemRequestID := d.Id()
	// Get client handle
	p := readProviderConfiguration(d)

	requestTemplate, validityErr := p.checkConfigValuesValidity(d)
	if validityErr != nil {
		return validityErr
	}
	validityErr = p.checkResourceConfigValidity(requestTemplate)
	if validityErr != nil {
		return validityErr
	}

	resourceActions, err := vraClient.GetResourceActions(catalogItemRequestID)
	if err != nil {
		log.Errorf("Error while reading resource actions for the request %v: %v ", catalogItemRequestID, err.Error())
		return fmt.Errorf("Error while reading resource actions for the request %v: %v  ", catalogItemRequestID, err.Error())
	}

	// If any change made in resource_configuration.
	if d.HasChange("resource_configuration") {
		for _, resources := range resourceActions.Content {
			if resources.ResourceTypeRef.ID == sdk.InfrastructureVirtual {
				var reconfigureEnabled bool
				var reconfigureActionID string
				for _, op := range resources.Operations {
					if op.Name == sdk.Reconfigure {
						reconfigureEnabled = true
						reconfigureActionID = op.OperationID
						break
					}
				}
				// if reconfigure action is not available for any resource of the deployment
				// return with an error message
				if !reconfigureEnabled {
					return fmt.Errorf("Update is not allowed for resource %v, your entitlement has no Reconfigure action enabled", resources.ID)
				}
				resourceData := resources.ResourceData
				entries := resourceData.Entries
				for _, entry := range entries {
					if entry.Key == sdk.Component {
						entryValue := entry.Value
						var componentName string
						for k, v := range entryValue {
							if k == "value" {
								componentName = v.(string)
							}
						}
						log.Info("Retrieving reconfigure action template for the component: %v ", componentName)

						resourceActionTemplate, err := vraClient.GetResourceActionTemplate(resources.ID, reconfigureActionID)
						if err != nil {
							log.Errorf("Error retrieving reconfigure action template for the component %v: %v ", componentName, err.Error())
							return fmt.Errorf("Error retrieving reconfigure action template for the component %v: %v ", componentName, err.Error())
						}
						configChanged := false
						for configKey := range p.ResourceConfiguration {
							var returnFlag bool
							//compare resource list (resource_name) with user configuration fields
							if strings.HasPrefix(configKey, componentName+".") {
								//If user_configuration contains resource_list element
								// then split user configuration key into resource_name and field_name
								nameList := strings.Split(configKey, componentName+".")
								//actionResponseInterface := actionResponse.(map[string]interface{})
								//Function call which changes the template field values with  user values
								//Replace existing values with new values in resource child template
								resourceActionTemplate.Data, returnFlag = utils.ReplaceValueInRequestTemplate(
									resourceActionTemplate.Data,
									nameList[1],
									p.ResourceConfiguration[configKey])
								if returnFlag {
									configChanged = true
								}
							}
						}
						oldData, _ := d.GetChange("resource_configuration")
						// If template value got changed then set post call and update resource child
						if configChanged {
							// This request id is for the reconfigure action on this machine and
							// will be used to track the status of the reconfigure request for this resource.
							// It will not replace the initial catalog item request id
							requestID, err := vraClient.PostResourceAction(resources.ID, reconfigureActionID, resourceActionTemplate)
							if err != nil {
								err = d.Set("resource_configuration", oldData)
								if err != nil {
									return err
								}
								log.Errorf("The update request failed with error: %v ", err)
								return err
							}
							status, err := waitForRequestCompletion(d, meta, requestID)
							if err != nil {
								// if the update request fails, go back to the old state and return the error
								if status == sdk.Failed {
									setErr := d.Set("resource_configuration", oldData)
									if setErr != nil {
										return setErr
									}
								}
								return err
							}
						}
					}
				}
			}
		}
	}
	return resourceVra7DeploymentRead(d, meta)
}

// Terraform call - terraform refresh
// This function retrieves the latest state of a vRA 7 deployment. Terraform updates its state based on
// the information returned by this function.
func resourceVra7DeploymentRead(d *schema.ResourceData, meta interface{}) error {
	vraClient = meta.(*sdk.APIClient)
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
		return fmt.Errorf("Resource view failed to load:  %v", errTemplate)
	}

	resourceDataMap := make(map[string]map[string]interface{})
	for _, resource := range requestResourceView.Content {
		if resource.ResourceType == sdk.InfrastructureVirtual {
			resourceData := resource.ResourcesData
			log.Info("The resource data map of the resource %v is: \n%v", resourceData.Component, resource.ResourcesData)
			dataVals := make(map[string]interface{})
			resourceDataMap[resourceData.Component] = dataVals
			dataVals[sdk.MachineCPU] = resourceData.CPU
			dataVals[sdk.MachineStorage] = resourceData.Storage
			dataVals[sdk.IPAddress] = resourceData.IPAddress
			dataVals[sdk.MachineMemory] = resourceData.Memory
			dataVals[sdk.MachineName] = resourceData.MachineName
			dataVals[sdk.MachineGuestOs] = resourceData.MachineGuestOperatingSystem
			dataVals[sdk.MachineBpName] = resourceData.MachineBlueprintName
			dataVals[sdk.MachineType] = resourceData.MachineType
			dataVals[sdk.MachineReservationName] = resourceData.MachineReservationName
			dataVals[sdk.MachineInterfaceType] = resourceData.MachineInterfaceType
			dataVals[sdk.MachineID] = resourceData.MachineID
			dataVals[sdk.MachineGroupName] = resourceData.MachineGroupName
			dataVals[sdk.MachineDestructionDate] = resourceData.MachineDestructionDate
			// Handle Network Info
			for idx, netDetails := range resourceData.Networks {
				log.Info("The Network list value is for idx %i  = %+v", idx, netDetails)
				networkIndexName := "Network" + strconv.Itoa(idx)
				dataVals[networkIndexName+".IPAddress"] = netDetails.NetworkAddressInfo.IPAddress
				dataVals[networkIndexName+".MACAddress"] = netDetails.NetworkAddressInfo.MACAddress
				dataVals[networkIndexName+".Name"] = netDetails.NetworkAddressInfo.Name
			}

		}
	}
	resourceConfiguration, _ := d.Get("resource_configuration").(map[string]interface{})
	resourceConfiguration, changed := utils.UpdateResourceConfigurationMap(resourceConfiguration, resourceDataMap)

	if changed {
		setError := d.Set("resource_configuration", resourceConfiguration)
		if setError != nil {
			return fmt.Errorf(setError.Error())
		}
	}
	return nil
}

//Function use - To delete resources which are created by terraform and present in state file
//Terraform call - terraform destroy
func resourceVra7DeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	vraClient = meta.(*sdk.APIClient)
	//Get requester machine ID from schema.dataresource
	catalogItemRequestID := d.Id()
	// Throw an error if request ID has no value or empty value
	if len(d.Id()) == 0 {
		return fmt.Errorf("Resource not found")
	}
	log.Info("Calling delete resource for the request id %v ", catalogItemRequestID)

	resourceView, err := vraClient.GetRequestResourceView(catalogItemRequestID)
	if err != nil {
		return fmt.Errorf("Resource view failed to load:  %v", err)
	}
	if len(resourceView.Content) == 0 {
		//If resource gets deleted then unset the resource ID from state file
		d.SetId("")
		return fmt.Errorf("The resource cannot be found")
	}

	resourceActions, err := vraClient.GetResourceActions(catalogItemRequestID)
	if err != nil {
		return err
	}

	for _, resources := range resourceActions.Content {
		if resources.ResourceTypeRef.ID == sdk.DeploymentResourceType {
			deploymentName := resources.Name
			var destroyEnabled bool
			var destroyActionID string
			for _, op := range resources.Operations {
				if op.Name == sdk.Destroy {
					destroyEnabled = true
					destroyActionID = op.OperationID
					break
				}
			}
			// if destroy deployment action is not available for the deployment
			// return with an error message
			if !destroyEnabled {
				return fmt.Errorf("The deployment %v cannot be destroyed, your entitlement has no Destroy Deployment action enabled", deploymentName)
			}
			resourceActionTemplate, err := vraClient.GetResourceActionTemplate(resources.ID, destroyActionID)
			if err != nil {
				log.Errorf(DestroyActionTemplateError, deploymentName, err.Error())
				return fmt.Errorf(DestroyActionTemplateError, deploymentName, err.Error())
			}
			requestID, err := vraClient.PostResourceAction(resources.ID, destroyActionID, resourceActionTemplate)
			if err != nil {
				log.Errorf("The destroy deployment request failed with error: %v ", err)
				return err
			}
			status, err := waitForRequestCompletion(d, meta, requestID)
			if err != nil {
				if status == sdk.Successful {
					d.SetId("")
				}
			}
		}
	}
	return nil
}

// check if the resource configuration is valid in the terraform config file
func (p *ProviderSchema) checkResourceConfigValidity(requestTemplate *sdk.CatalogItemRequestTemplate) error {
	log.Info("Checking if the terraform config file is valid")

	// Get all component names in the blueprint corresponding to the catalog item.
	componentSet := make(map[string]bool)
	for field := range requestTemplate.Data {
		if reflect.ValueOf(requestTemplate.Data[field]).Kind() == reflect.Map {
			componentSet[field] = true
		}
	}
	log.Info("The component name(s) in the blueprint corresponding to the catalog item: %v\n", componentSet)

	var invalidKeys []string
	// check if the keys in the resourceConfiguration map exists in the componentSet
	// if the key in config is machine1.vsphere.custom.location, match every string after each dot
	// until a matching string is found in componentSet.
	// If found, it's a valid key else the component name is invalid
	for k := range p.ResourceConfiguration {
		var key = k
		var isValid bool
		for strings.LastIndex(key, ".") != -1 {
			lastIndex := strings.LastIndex(key, ".")
			key = key[0:lastIndex]
			if _, ok := componentSet[key]; ok {
				log.Info("The component name %s in the terraform config file is valid ", key)
				isValid = true
				break
			}
		}
		if !isValid {
			invalidKeys = append(invalidKeys, k)
		}
	}
	// there are invalid resource config keys in the terraform config file, abort and throw an error
	if len(invalidKeys) > 0 {
		log.Error("The resource_configuration in the config file has invalid component name(s): %v ", strings.Join(invalidKeys, ", "))
		return fmt.Errorf(ConfigInvalidError, strings.Join(invalidKeys, ", "))
	}
	return nil
}

// check if the values provided in the config file are valid and set
// them in the resource schema. Requires to call APIs
func (p *ProviderSchema) checkConfigValuesValidity(d *schema.ResourceData) (*sdk.CatalogItemRequestTemplate, error) {
	// 	// If catalog_name and catalog_id both not provided then return an error
	if len(p.CatalogItemName) <= 0 && len(p.CatalogItemID) <= 0 {
		return nil, fmt.Errorf("Either catalog_name or catalog_id should be present in given configuration")
	}

	var catalogItemIDFromName string
	var catalogItemNameFromID string
	var err error
	// if catalog item id is provided, fetch the catalog item name
	if len(p.CatalogItemName) > 0 {
		catalogItemIDFromName, err = vraClient.ReadCatalogItemByName(p.CatalogItemName)
		if err != nil || catalogItemIDFromName == "" {
			return nil, fmt.Errorf("Error in finding catalog item id corresponding to the catlog item name %v: \n %v", p.CatalogItemName, err)
		}
		log.Info("The catalog item id provided in the config is %v\n", catalogItemIDFromName)
	}

	// if catalog item name is provided, fetch the catalog item id
	if len(p.CatalogItemID) > 0 { // else if both are provided and matches or just id is provided, use id
		catalogItemNameFromID, err = vraClient.ReadCatalogItemNameByID(p.CatalogItemID)
		if err != nil || catalogItemNameFromID == "" {
			return nil, fmt.Errorf("Error in finding catalog item name corresponding to the catlog item id %v: \n %v", p.CatalogItemID, err)
		}
		log.Info("The catalog item name corresponding to the catalog item id in the config is:  %v\n", catalogItemNameFromID)
	}

	// if both catalog item name and id are provided but does not belong to the same catalog item, throw an error
	if len(p.CatalogItemName) > 0 && len(p.CatalogItemID) > 0 && (catalogItemIDFromName != p.CatalogItemID || catalogItemNameFromID != p.CatalogItemName) {
		log.Error(CatalogItemIDNameNotMatchingErr, p.CatalogItemName, p.CatalogItemID)
		return nil, fmt.Errorf(CatalogItemIDNameNotMatchingErr, p.CatalogItemName, p.CatalogItemID)
	} else if len(p.CatalogItemID) > 0 { // else if both are provided and matches or just id is provided, use id
		d.Set("catalog_item_id", p.CatalogItemID)
		d.Set("catalog_item_name", catalogItemNameFromID)
	} else if len(p.CatalogItemName) > 0 { // else if name is provided, use the id fetched from the name
		d.Set("catalog_item_id", catalogItemIDFromName)
		d.Set("catalog_item_name", p.CatalogItemName)
	}

	// update the catalogItemID var with the updated id
	p.CatalogItemID = d.Get("catalog_item_id").(string)

	// Get request template for catalog item.
	requestTemplate, err := vraClient.GetCatalogItemRequestTemplate(p.CatalogItemID)
	if err != nil {
		return nil, err
	}
	log.Info("The request template data corresponding to the catalog item %v is: \n %v\n", p.CatalogItemID, requestTemplate.Data)

	updateRequestTemplateFromDeploymentConfiguration(p.DeploymentConfiguration, requestTemplate)

	// get the business group id from name
	var businessGroupIDFromName string
	if len(p.BusinessGroupName) > 0 {
		businessGroupIDFromName, err = vraClient.GetBusinessGroupID(p.BusinessGroupName, vraClient.Tenant)
		if err != nil || businessGroupIDFromName == "" {
			return nil, err
		}
	}

	//if both business group name and id are provided but does not belong to the same business group, throw an error
	if len(p.BusinessGroupName) > 0 && len(p.BusinessGroupID) > 0 && businessGroupIDFromName != p.BusinessGroupID {
		log.Error(BusinessGroupIDNameNotMatchingErr, p.BusinessGroupName, p.BusinessGroupID)
		return nil, fmt.Errorf(BusinessGroupIDNameNotMatchingErr, p.BusinessGroupName, p.BusinessGroupID)
	} else if len(p.BusinessGroupID) > 0 { // else if both are provided and matches or just id is provided, use id
		log.Info("Setting business group id %s ", p.BusinessGroupID)
		requestTemplate.BusinessGroupID = p.BusinessGroupID
	} else if len(p.BusinessGroupName) > 0 { // else if name is provided, use the id fetched from the name
		log.Info("Setting business group id %s for the group %s ", businessGroupIDFromName, p.BusinessGroupName)
		requestTemplate.BusinessGroupID = businessGroupIDFromName
	}
	return requestTemplate, nil
}

func updateRequestTemplateFromDeploymentConfiguration(deploymentConfiguration map[string]interface{}, requestTemplate *sdk.CatalogItemRequestTemplate) {
	for field := range deploymentConfiguration {
		value := deploymentConfiguration[field]

		requestTemplate.Data[field] = utils.UnmarshalJsonStringIfNecessary(field, value)
	}
}

// check the request status on apply and update
func waitForRequestCompletion(d *schema.ResourceData, meta interface{}, requestID string) (string, error) {

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
			log.Error("Request Failed with message %v ", reqestStatusView.RequestCompletion.CompletionDetails)
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

// read the config file
func readProviderConfiguration(d *schema.ResourceData) *ProviderSchema {

	log.Info("Reading the provider configuration data.....")
	providerSchema := ProviderSchema{
		CatalogItemName:         strings.TrimSpace(d.Get("catalog_item_name").(string)),
		CatalogItemID:           strings.TrimSpace(d.Get("catalog_item_id").(string)),
		Description:             strings.TrimSpace(d.Get("description").(string)),
		Reasons:                 strings.TrimSpace(d.Get("reasons").(string)),
		BusinessGroupName:       strings.TrimSpace(d.Get("businessgroup_name").(string)),
		BusinessGroupID:         strings.TrimSpace(d.Get("businessgroup_id").(string)),
		WaitTimeout:             d.Get("wait_timeout").(int) * 60,
		FailedMessage:           strings.TrimSpace(d.Get("failed_message").(string)),
		ResourceConfiguration:   d.Get("resource_configuration").(map[string]interface{}),
		DeploymentConfiguration: d.Get("deployment_configuration").(map[string]interface{}),
	}

	log.Info("The values provided in the TF config file is: \n %v ", providerSchema)
	return &providerSchema
}
