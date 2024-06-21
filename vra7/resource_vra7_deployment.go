package vra7

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	logging "github.com/op/go-logging"
	"github.com/vmware/terraform-provider-vra7/sdk"
	"github.com/vmware/terraform-provider-vra7/utils"
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
	DeploymentDestroyAction string
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
			"deployment_destroy_action": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Destroy",
			},
			"resource_configuration": resourceConfigurationSchema(),
			"lease_days": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"expiry_date": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"request_status": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
	if p.Lease != 0 {
		requestTemplate.Data["_leaseDays"] = p.Lease
	}
	for field, value := range p.DeploymentConfiguration {
		requestTemplate.Data[field] = utils.UnmarshalJSONStringIfNecessary(field, value)
	}

	for _, rConfig := range p.ResourceConfiguration {
		tempConfigMap := make(map[string]interface{})
		for index, element := range rConfig.Configuration {
			tempConfigMap[index] = element
		}
		if rConfig.Cluster != 0 {
			tempConfigMap["_cluster"] = rConfig.Cluster
		}
		for propertyName, propertyValue := range tempConfigMap {
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
	replaced := ReplaceValueInRequestTemplate(templateInterface, field, value)

	if !replaced {
		templateInterface["data"] = AddValueToRequestTemplate(templateInterface["data"].(map[string]interface{}), field, value)
	}
	return templateInterface
}

// This function updates the state of a vRA 7 Deployment when changes to a Terraform file are applied.
// The update is performed on the Deployment using supported (day-2) actions.
func resourceVra7DeploymentUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Info("Updating the resource vra7_deployment with request id %s", d.Id())
	vraClient := meta.(*sdk.APIClient)

	p, err := readProviderConfiguration(d, vraClient)
	if err != nil {
		return err
	}
	_, validityErr := p.checkResourceConfigValidity(vraClient)
	if validityErr != nil {
		return validityErr
	}

	// Change Lease Day 2 operation
	if d.HasChange("expiry_date") {
		deploymentResourceActions, err := vraClient.GetResourceActions(p.DeploymentID)
		if err != nil {
			return err
		}
		deploymentActionsMap := GetActionNameIDMap(deploymentResourceActions)
		changeLeaseActionID := deploymentActionsMap["Change Lease"]
		if changeLeaseActionID != "" {
			resourceActionTemplate, _ := vraClient.GetResourceActionTemplate(p.DeploymentID, changeLeaseActionID)
			log.Info("Starting Change Lease action on the deployment with id %v. The lease will be extended by %v days.", p.DeploymentID, p.Lease)
			_ = ReplaceValueInRequestTemplate(
				resourceActionTemplate.Data, "provider-ExpirationDate", d.Get("expiry_date").(string))
			resourceActionTemplate.Description = d.Get("description").(string)
			resourceActionTemplate.Reasons = d.Get("reasons").(string)
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
			index, oldResourceConfig := GetResourceConfigurationByComponent(oldResourceConfigList, newResourceConfig.ComponentName)
			if index != -1 {

				deploymentResourceActions, err := vraClient.GetResourceActions(p.DeploymentID)
				if err != nil {
					return err
				}
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
						resourceActionTemplate.Description = d.Get("description").(string)
						resourceActionTemplate.Reasons = d.Get("reasons").(string)
						// get the map from the action template corresponding to the key which is the component name
						actionTemplateDataMap := GetActionTemplateDataByComponent(resourceActionTemplate.Data, newResourceConfig.ComponentName)
						// update the template with the new cluster size
						log.Info("Starting Scale Out action on the deployment with id %v for the component %v. The cluster size will be increased from %v to %v.",
							p.DeploymentID, oldResourceConfig.ComponentName, oldResourceConfig.Cluster, newResourceConfig.Cluster)
						_ = ReplaceValueInRequestTemplate(
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
						log.Info("Successfully completed the Scale Out action for the deployment with id %v.", p.DeploymentID)
					} else if oldResourceConfig.Cluster > newResourceConfig.Cluster && deploymentActionsMap[sdk.ScaleIn] != "" {
						// Scale In Day 2 operation
						scaleInActionID := deploymentActionsMap[sdk.ScaleIn]
						// get the action template for scale in
						resourceActionTemplate, err := vraClient.GetResourceActionTemplate(p.DeploymentID, scaleInActionID)
						if err != nil {
							return err
						}
						resourceActionTemplate.Description = d.Get("description").(string)
						resourceActionTemplate.Reasons = d.Get("reasons").(string)
						// get the map from the action template corresponding to the key which is the component name
						actionTemplateDataMap := GetActionTemplateDataByComponent(resourceActionTemplate.Data, newResourceConfig.ComponentName)
						// update the template with the new cluster size
						log.Info("Starting Scale In action on the deployment with id %v for the component %v. The cluster size will be decresed from %v to %v.",
							p.DeploymentID, oldResourceConfig.ComponentName, oldResourceConfig.Cluster, newResourceConfig.Cluster)
						_ = ReplaceValueInRequestTemplate(
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
						log.Info("Successfully completed the Scale In action for the deployment with id %v.", p.DeploymentID)
					}
				}
			}
		}

		for _, newRC := range newResourceConfigList {
			cName := newRC.ComponentName
			index, oldRC := GetResourceConfigurationByComponent(oldResourceConfigList, cName)

			if index != -1 {
				newConfig := newRC.Configuration
				for _, instance := range oldRC.Instances {
					vmResourceActions, err := vraClient.GetResourceActions(instance.ResourceID)
					if err != nil {
						return err
					}
					vmResourceActionsMap := GetActionNameIDMap(vmResourceActions)
					if vmResourceActionsMap[sdk.Reconfigure] != "" {
						reconfigureActionID := vmResourceActionsMap[sdk.Reconfigure]
						resourceActionTemplate, _ := vraClient.GetResourceActionTemplate(instance.ResourceID, reconfigureActionID)
						resourceActionTemplate.Description = d.Get("description").(string)
						resourceActionTemplate.Reasons = d.Get("reasons").(string)
						configChanged := false
						actionTemplateDataMap := resourceActionTemplate.Data
						// checking if any property has changed in the new configuration
						for propertyName, propertyValue := range newConfig {
							if oldRC.Configuration[propertyName] != propertyValue {
								_ = ReplaceValueInRequestTemplate(
									actionTemplateDataMap, propertyName, propertyValue)
								if !configChanged {
									configChanged = true
								}
							}
						}
						if configChanged {
							log.Info("Starting Reconfigure action on the component %v.", cName)
							requestID, err := vraClient.PostResourceAction(instance.ResourceID, reconfigureActionID, resourceActionTemplate)
							if err != nil {
								log.Errorf("The reconfigure request failed with error: %v ", err)
								return err
							}
							log.Info("The Reconfigure operation for the component %v has been submitted", cName)
							_, err = waitForRequestCompletion(d, meta, requestID)
							if err != nil {
								log.Errorf("The reconfigure request for component %v failed with error: %v ", cName, err)
								return err
							}
							log.Info("Successfully completed the Reconfigure action on the component %v.", cName)
						}
					}
				}
			}
		}
	}

	// the description and reasons cannot be updated without any valid day-2 opearation
	if (d.HasChange("description") || d.HasChange("reasons")) && (!d.HasChange("lease_days") && !d.HasChange("resource_configuration")) {
		return fmt.Errorf("Updating only description and/or reasons is not supported. You can update them during any supported Day-2 actions")
	}

	log.Info("Finished updating the resource vra7_deployment with request id %s", d.Id())
	return resourceVra7DeploymentRead(d, meta)
}

// This function retrieves the latest state of a vRA 7 deployment. Terraform updates its state based on
// the information returned by this function.
func resourceVra7DeploymentRead(d *schema.ResourceData, meta interface{}) error {

	log.Info("Reading the resource vra7_deployment with request id %s ", d.Id())
	vraClient := meta.(*sdk.APIClient)

	p, err := readProviderConfiguration(d, vraClient)
	if err != nil {
		return err
	}

	// Get the ID of the catalog request that was used to provision this Deployment. This id
	// will remain the same for this deployment across any actions on the machines like reconfigure, etc.
	catalogItemRequestID := d.Id()

	deploymentID, err := vraClient.GetDeploymentIDFromRequest(catalogItemRequestID)
	if err != nil {
		return err
	}
	// Since the resource view API above do not provide the cluster value, it is calculated
	// by tracking the component name and updated in the state file
	clusterCountMap := make(map[string]int)
	// parse the resource view API response and create a resource configuration list that will contain information
	// of the deployed VMs
	var resourceConfigList []sdk.ResourceConfigurationStruct

	deployment, err := vraClient.GetDeployment(deploymentID)

	if err != nil {
		return err
	}

	d.Set("catalog_item_id", deployment.CatalogItem.ID)
	d.Set("catalog_item_name", deployment.CatalogItem.Label)
	d.Set("deployment_id", deploymentID)
	d.Set("description", deployment.Description)
	d.Set("created_date", deployment.CreatedDate)
	d.Set("expiry_date", deployment.ExpiryDate)
	d.Set("name", deployment.Name)
	d.Set("businessgroup_id", deployment.Subtenant.ID)
	d.Set("businessgroup_name", deployment.Subtenant.Label)

	owners := make([]map[string]string, 0)
	for _, owner := range deployment.Owners {
		ownerMap := make(map[string]string)
		ownerMap["id"] = owner.ID
		ownerMap["name"] = owner.Name
		owners = append(owners, ownerMap)
	}
	d.Set("owners", owners)

	for _, component := range deployment.Components {

		if component.Type == sdk.InfrastructureVirtual {

			data := component.Data
			componentName := data["Component"].(string)

			if componentName != "" {
				instance := sdk.Instance{}
				instance.IPAddress = data["ip_address"].(string)
				instance.Name = component.Name
				instance.ResourceID = component.ID
				instance.ResourceType = component.Type
				instance.Properties = data

				// checking to see if a resource configuration struct exists for the component name
				// if yes, then add another instance to the instances list of that resource config struct
				// at index of resource config list
				// else create a new rescource config struct and add to the resource config list
				index, rcStruct := GetResourceConfigurationByComponent(resourceConfigList, componentName)

				if index == -1 {
					rcStruct.ComponentName = componentName
					rcStruct.RequestID = catalogItemRequestID
					rcStruct.ParentResourceID = component.ParentID
					if p != nil && p.ResourceConfiguration != nil {
						rcStruct.Configuration = GetConfiguration(componentName, p.ResourceConfiguration)
					}
					rcStruct.Instances = make([]sdk.Instance, 0)
					rcStruct.Instances = append(rcStruct.Instances, instance)
					resourceConfigList = append(resourceConfigList, rcStruct)
				} else {
					rcStruct.Instances = append(rcStruct.Instances, instance)
					resourceConfigList[index] = rcStruct
				}
				clusterCountMap[componentName] = clusterCountMap[componentName] + 1
			}
		}
	}

	if err := d.Set("resource_configuration", flattenResourceConfigurations(resourceConfigList, clusterCountMap)); err != nil {
		return fmt.Errorf("error setting resource configuration - error: %v", err)
	}

	log.Info("Finished reading the resource vra7_deployment with request id %s", d.Id())
	return nil
}

// Function use - To delete resources which are created by terraform and present in state file
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
	destroyActionID := deploymentActionsMap[p.DeploymentDestroyAction]
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
		DeploymentDestroyAction: d.Get("deployment_destroy_action").(string),
		DeploymentConfiguration: d.Get("deployment_configuration").(map[string]interface{}),
	}

	// if catalog item name is provided, fetch the catalog item id
	if len(providerSchema.CatalogItemID) == 0 && len(providerSchema.CatalogItemName) > 0 {
		id, err := vraClient.ReadCatalogItemByName(providerSchema.CatalogItemName)
		if err != nil {
			return &providerSchema, err
		}
		providerSchema.CatalogItemID = id
	}

	// get the business group id from name
	if len(providerSchema.BusinessGroupID) == 0 && len(providerSchema.BusinessGroupName) > 0 {
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
	sleepFor := 20
	status := ""
	for i := 0; i < waitTimeout/sleepFor; i++ {
		log.Info("Waiting for %d seconds before checking request status.", sleepFor)
		time.Sleep(time.Duration(sleepFor) * time.Second)
		requestStatusView, _ := vraClient.GetRequestStatus(requestID)
		status = requestStatusView.Phase
		d.Set("request_status", status)
		log.Info("Checking to see the status of the request. Status: %s.", status)
		if status == sdk.Successful {
			log.Info("Request is SUCCESSFUL.")
			return sdk.Successful, nil
		} else if status == sdk.Failed {
			return sdk.Failed, fmt.Errorf("Request failed \n %v ", requestStatusView.RequestCompletion.CompletionDetails)
		} else if status == sdk.Rejected {
			return sdk.Rejected, fmt.Errorf("Request rejected \n %v ", requestStatusView.RequestCompletion.CompletionDetails)
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
func GetResourceConfigurationByComponent(resourceConfigurationList []sdk.ResourceConfigurationStruct, component string) (int, sdk.ResourceConfigurationStruct) {
	for index, rConfig := range resourceConfigurationList {
		if rConfig.ComponentName == component {
			return index, rConfig
		}
	}
	return -1, sdk.ResourceConfigurationStruct{}
}

// GetActionNameIDMap returns a map of Action name and id
func GetActionNameIDMap(resourceActions []sdk.Operation) map[string]string {
	actionNameIDMap := make(map[string]string)
	for _, action := range resourceActions {
		actionNameIDMap[action.Name] = action.ID
	}
	return actionNameIDMap
}

// GetResourceByID return the resource config struct object filtered by ID
func GetResourceByID(resourceConfigStructList []sdk.ResourceConfigurationStruct, resourceID string) sdk.ResourceConfigurationStruct {
	var resourceConfigStruct sdk.ResourceConfigurationStruct
	for _, resourceStruct := range resourceConfigStructList {
		for _, instance := range resourceStruct.Instances {
			if instance.ResourceID == resourceID {
				resourceConfigStruct = resourceStruct
			}
		}
	}
	return resourceConfigStruct
}
