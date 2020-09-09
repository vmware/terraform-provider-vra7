package vra7

import (
	"fmt"
	"time"

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

func dataSourceVra7DeploymentRead(d *schema.ResourceData, meta interface{}) error {
	vraClient := meta.(*sdk.APIClient)

	id, idOk := d.GetOk("id")
	deploymentID, deploymentIDOk := d.GetOk("deployment_id")

	if !idOk && !deploymentIDOk {
		return fmt.Errorf("One of id or deployment_id must be assigned")
	}

	if deploymentID.(string) != "" {
		resource, err := vraClient.GetResource(deploymentID.(string))
		if err != nil {
			return err
		}
		id = resource.RequestID
	}

	// Since the resource view API above do not provide the cluster value, it is calculated
	// by tracking the component name and updated in the state file
	clusterCountMap := make(map[string]int)
	// parse the resource view API response and create a resource configuration list that will contain information
	// of the deployed VMs
	var resourceConfigList []sdk.ResourceConfigurationStruct

	currentPage := 1
	totalPages := 1

	for currentPage <= totalPages {
		requestResourceView, errTemplate := vraClient.GetRequestResourceView(id.(string), currentPage)
		// if resource does not exists, then unset the resource ID from state file
		if requestResourceView != nil && len(requestResourceView.Content) == 0 {
			d.SetId("")
			return fmt.Errorf("The resource cannot be found")
		}
		if errTemplate != nil || len(requestResourceView.Content) == 0 {
			return fmt.Errorf("Resource view failed to load with the error %v", errTemplate)
		}

		currentPage = requestResourceView.MetaData.Number + 1
		totalPages = requestResourceView.MetaData.TotalPages

		for _, resource := range requestResourceView.Content {

			// map containing the content of a resourceView response
			rMap := resource.(map[string]interface{})
			// fetching the catalog item request specific data
			requestID := rMap["requestId"].(string)
			requestState := rMap["requestState"].(string)
			// fetching common attributes of a resource. A resource can be Infrastructure.Virtual or a deployment, etc
			resourceType := rMap["resourceType"].(string)
			dateCreated := rMap["dateCreated"].(string)
			lastUpdated := rMap["lastUpdated"].(string)
			resourceID := rMap["resourceId"].(string)
			name := rMap["name"].(string)
			description := ""
			status := ""
			if _, ok := rMap["status"]; !ok {
				status = rMap["status"].(string)
			}
			if _, ok := rMap["description"]; !ok {
				description = rMap["description"].(string)
			}

			// if the resource type is VMs, update the resource_configuration attribute
			if resourceType == sdk.InfrastructureVirtual {
				data := rMap["data"].(map[string]interface{})
				componentName := data["Component"].(string)
				if componentName != "" {
					instance := sdk.Instance{}
					instance.DateCreated = dateCreated
					instance.LastUpdated = lastUpdated
					instance.IPAddress = data["ip_address"].(string)
					instance.Name = name
					instance.ResourceID = resourceID
					instance.ResourceType = resourceType
					instance.Properties = data
					instance.Description = description
					instance.Status = status
					componentName := data["Component"].(string)

					// checking to see if a resource configuration struct exists for the component name
					// if yes, then add another instance to the instances list of that resource config struct
					// at index of resource config list
					// else create a new rescource config struct and add to the resource config list
					index, rcStruct := GetResourceConfigurationByComponent(resourceConfigList, componentName)

					if index == -1 {
						rcStruct.ComponentName = componentName
						rcStruct.RequestID = requestID
						rcStruct.RequestState = requestState
						rcStruct.ParentResourceID = rMap["parentResourceId"].(string)
						rcStruct.Instances = make([]sdk.Instance, 0)
						rcStruct.Instances = append(rcStruct.Instances, instance)
						resourceConfigList = append(resourceConfigList, rcStruct)
					} else {
						rcStruct.Instances = append(rcStruct.Instances, instance)
						resourceConfigList[index] = rcStruct
					}
					clusterCountMap[componentName] = clusterCountMap[componentName] + 1
				}

			} else if resourceType == sdk.DeploymentResourceType {
				d.Set("catalog_item_id", rMap["catalogItemId"].(string))
				d.Set("catalog_item_name", rMap["catalogItemLabel"].(string))
				d.Set("deployment_id", resourceID)
				d.Set("date_created", dateCreated)
				d.Set("last_updated", lastUpdated)
				d.Set("tenant_id", rMap["tenantId"].(string))
				d.Set("owners", rMap["owners"].([]interface{}))
				d.Set("name", name)
				d.Set("businessgroup_id", rMap["businessGroupId"].(string))
				d.Set("description", description)
				d.Set("request_status", status)
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
			}
		}
	}

	if err := d.Set("resource_configuration", flattenResourceConfigurations(resourceConfigList, clusterCountMap)); err != nil {
		return fmt.Errorf("error setting resource configuration - error: %v", err)
	}
	d.SetId(id.(string))

	log.Info("Finished reading the resource vra7_deployment with request id %s", d.Id())
	return nil
}
