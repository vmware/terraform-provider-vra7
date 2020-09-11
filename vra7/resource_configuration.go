package vra7

import (
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/terraform-provider-vra7/sdk"
)

func resourceConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"component_name": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"configuration": {
					Type:     schema.TypeMap,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"cluster": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  1,
				},
				"parent_resource_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ip_address": {
					Type:     schema.TypeString,
					Computed: true,
					Removed:  "The ip_address is removed from here and available under instances",
				},
				"request_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"request_state": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"instances": instancesSchema(),
			},
		},
	}
}

func dataResourceConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"component_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"configuration": {
					Type:     schema.TypeMap,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"cluster": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"parent_resource_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ip_address": {
					Type:     schema.TypeString,
					Computed: true,
					Removed:  "The ip_address is removed from here and available under instances",
				},
				"request_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"request_state": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"instances": instancesSchema(),
			},
		},
	}
}

func instancesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"resource_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ip_address": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"resource_type": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"date_created": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"last_updated": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"properties": {
					Type:     schema.TypeMap,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func expandResourceConfiguration(rConfigurations []interface{}) []sdk.ResourceConfigurationStruct {
	configs := make([]sdk.ResourceConfigurationStruct, 0, len(rConfigurations))
	for _, config := range rConfigurations {
		configMap := config.(map[string]interface{})
		instances := make([]sdk.Instance, 0)
		for _, i := range configMap["instances"].([]interface{}) {
			ins := i.(map[string]interface{})
			instance := sdk.Instance{
				ResourceID:   ins["resource_id"].(string),
				Name:         ins["name"].(string),
				IPAddress:    ins["ip_address"].(string),
				ResourceType: ins["resource_type"].(string),
				Status:       ins["status"].(string),
				Description:  ins["description"].(string),
				DateCreated:  ins["date_created"].(string),
				LastUpdated:  ins["last_updated"].(string),
				Properties:   ins["properties"].(map[string]interface{}),
			}
			instances = append(instances, instance)
		}
		rConfig := sdk.ResourceConfigurationStruct{
			ComponentName:    configMap["component_name"].(string),
			Configuration:    configMap["configuration"].(map[string]interface{}),
			Cluster:          configMap["cluster"].(int),
			ParentResourceID: configMap["parent_resource_id"].(string),
			RequestID:        configMap["request_id"].(string),
			RequestState:     configMap["request_state"].(string),
			Instances:        instances,
		}
		configs = append(configs, rConfig)
	}
	return configs
}

func flattenResourceConfigurations(resourceConfigList []sdk.ResourceConfigurationStruct, clusterCountMap map[string]int) []map[string]interface{} {
	if len(resourceConfigList) == 0 {
		return make([]map[string]interface{}, 0)
	}
	rConfigs := make([]map[string]interface{}, 0, len(resourceConfigList))
	for _, config := range resourceConfigList {
		helper := make(map[string]interface{})
		instances := make([]map[string]interface{}, 0)
		for _, instance := range config.Instances {
			instanceMap := make(map[string]interface{})
			instanceMap["resource_id"] = instance.ResourceID
			instanceMap["resource_type"] = instance.ResourceType
			instanceMap["status"] = instance.Status
			instanceMap["name"] = instance.Name
			instanceMap["date_created"] = instance.DateCreated
			instanceMap["ip_address"] = instance.IPAddress
			propMap, configurationMap := parseDataMap(instance.Properties, config.Configuration)
			instanceMap["properties"] = propMap
			instances = append(instances, instanceMap)
			helper["configuration"] = configurationMap
		}
		helper["instances"] = instances
		helper["component_name"] = config.ComponentName
		helper["request_id"] = config.RequestID
		helper["parent_resource_id"] = config.ParentResourceID
		helper["request_state"] = config.RequestState
		helper["cluster"] = clusterCountMap[config.ComponentName]

		rConfigs = append(rConfigs, helper)
	}

	return rConfigs
}

func parseDataMap(resourceData map[string]interface{}, configurationMap map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	stateMap := make(map[string]interface{})
	resourcePropertyMapper := ResourceMapper()
	for key, value := range resourceData {

		if i, ok := resourcePropertyMapper[key]; ok {
			key = i
		}
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice:
			parseArray(key, stateMap, configurationMap, value.([]interface{}))
		case reflect.Map:
			parseMap(key, stateMap, configurationMap, value.(map[string]interface{}))
		default:
			stateMap[key] = convToString(value)
			if _, ok := configurationMap[key]; ok {
				configurationMap[key] = convToString(value)
			}
		}
	}
	return stateMap, configurationMap
}

func parseMap(prefix string, stateMap map[string]interface{}, configurationMap map[string]interface{}, data map[string]interface{}) {

	for key, value := range data {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Slice:
			parseArray(prefix+"."+key, stateMap, configurationMap, value.([]interface{}))
		case reflect.Map:
			parseMap(prefix+"."+key, stateMap, configurationMap, value.(map[string]interface{}))
		default:
			stateMap[prefix+"."+key] = convToString(value)
			if _, ok := configurationMap[prefix+"."+key]; ok {
				configurationMap[key] = convToString(value)
			}
		}
	}
}

func parseArray(prefix string, stateMap map[string]interface{}, configurationMap map[string]interface{}, value []interface{}) {

	for index, val := range value {
		v := reflect.ValueOf(val)
		switch v.Kind() {
		case reflect.Map:
			/* for properties like NETWORK_LIST, DISK_VOLUMES etc, the value is a slice of map as follows.
			Out of all the information, only data is important information, so leaving out rest of the properties
			 "NETWORK_LIST":[
					{
						"componentTypeId":"",
						"componentId":null,
						"classId":"",
						"typeFilter":null,
						"data":{
						   "NETWORK_MAC_ADDRESS":"00:50:56:b6:78:c6",
						   "NETWORK_NAME":"dvPortGroup-wdc-sdm-vm-1521"
						}
					 }
				  ]
			*/
			objMap := val.(map[string]interface{})
			for k, v := range objMap {
				if k == "data" {
					parseMap(prefix+"."+convToString(index), stateMap, configurationMap, v.(map[string]interface{}))
				}
			}
		default:
			stateMap[prefix+"."+convToString(index)] = convToString(val)
			if _, ok := configurationMap[prefix+"."+convToString(index)]; ok {
				configurationMap[prefix+"."+convToString(index)] = convToString(value)
			}
		}
	}
}

func convToString(value interface{}) string {

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return value.(string)
	case reflect.Float64:
		return strconv.FormatFloat(value.(float64), 'f', 0, 64)
	case reflect.Float32:
		return strconv.FormatFloat(value.(float64), 'f', 0, 32)
	case reflect.Int:
		return strconv.Itoa(value.(int))
	case reflect.Int32:
		return strconv.Itoa(value.(int))
	case reflect.Int64:
		return strconv.FormatInt(value.(int64), 10)
	case reflect.Bool:
		return strconv.FormatBool(value.(bool))
	}
	return ""
}
