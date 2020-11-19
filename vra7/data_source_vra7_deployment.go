package vra7

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/terraform-provider-vra7/sdk"
)

func dataSourceVra7Deployment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVra7DeploymentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"catalog_item_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_item_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reasons": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"businessgroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"businessgroup_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_configuration": dataResourceConfigurationSchema(),
			"lease_days": {
				Type:     schema.TypeInt,
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
			"expiry_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owners": {
				Type:     schema.TypeSet,
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

func dataSourceVra7DeploymentRead(d *schema.ResourceData, meta interface{}) error {
	vraClient := meta.(*sdk.APIClient)

	id, idOk := d.GetOk("id")
	deploymentID, deploymentIDOk := d.GetOk("deployment_id")

	if !idOk && !deploymentIDOk {
		return fmt.Errorf("One of id or deployment_id must be assigned")
	}

	//
	if id.(string) != "" {
		depID, err := vraClient.GetDeploymentIDFromRequest(id.(string))
		if err != nil {
			return err
		}
		deploymentID = depID
	}

	requestID := ""
	if deploymentID.(string) != "" {
		resource, err := vraClient.GetResource(deploymentID.(string))
		if err != nil {
			return err
		}
		requestID = resource.RequestID
	}

	// Since the resource view API above do not provide the cluster value, it is calculated
	// by tracking the component name and updated in the state file
	clusterCountMap := make(map[string]int)
	// parse the resource view API response and create a resource configuration list that will contain information
	// of the deployed VMs
	var resourceConfigList []sdk.ResourceConfigurationStruct

	deployment, err := vraClient.GetDeployment(deploymentID.(string))

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
					rcStruct.RequestID = id.(string)
					rcStruct.ParentResourceID = component.ParentID
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

	d.SetId(requestID)

	log.Info("Finished reading the data source vra7_deployment with request id %s", d.Id())
	return nil
}
