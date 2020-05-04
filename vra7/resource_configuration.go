package vra7

import (
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-vra7/sdk"
	"github.com/terraform-providers/terraform-provider-vra7/utils"
)

func resourceConfigurationSchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: !computed,
		Computed: computed,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"component_name": {
					Type:     schema.TypeString,
					Required: !computed,
					Computed: computed,
				},
				"configuration": {
					Type:     schema.TypeMap,
					Optional: true,
					Computed: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						if (old != "" && new == "") || (old == "" && new == "") {
							return true
						}
						return false
					},
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"cluster": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"resource_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"parent_resource_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ip_address": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"request_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"request_state": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"resource_type": {
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
			},
		},
	}
}

func expandResourceConfiguration(rConfigurations interface{}) []sdk.ResourceConfigurationStruct {
	configs := make([]sdk.ResourceConfigurationStruct, 0, len(rConfigurations.([]interface{})))

	for _, config := range rConfigurations.([]interface{}) {
		configMap := config.(map[string]interface{})

		rConfig := sdk.ResourceConfigurationStruct{
			ComponentName:    configMap["component_name"].(string),
			Configuration:    configMap["configuration"].(map[string]interface{}),
			Cluster:          configMap["cluster"].(int),
			Name:             configMap["name"].(string),
			Description:      configMap["description"].(string),
			DateCreated:      configMap["date_created"].(string),
			LastUpdated:      configMap["last_updated"].(string),
			ParentResourceID: configMap["parent_resource_id"].(string),
			ResourceID:       configMap["resource_id"].(string),
			ResourceType:     configMap["resource_type"].(string),
			RequestID:        configMap["request_id"].(string),
			RequestState:     configMap["request_state"].(string),
			IPAddress:        configMap["ip_address"].(string),
		}
		configs = append(configs, rConfig)
	}
	return configs
}

func flattenResourceConfigurations(configs []sdk.ResourceConfigurationStruct, clusterCountMap map[string]int) []map[string]interface{} {
	if len(configs) == 0 {
		return make([]map[string]interface{}, 0)
	}
	rConfigs := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		resourceDataMap := parseDataMap(config.Configuration)
		helper := make(map[string]interface{})
		helper["configuration"] = resourceDataMap
		helper["component_name"] = config.ComponentName
		helper["name"] = config.Name
		helper["date_created"] = config.DateCreated
		helper["last_updated"] = config.LastUpdated
		helper["resource_id"] = config.ResourceID
		helper["request_id"] = config.RequestID
		helper["parent_resource_id"] = config.ParentResourceID
		helper["status"] = config.Status
		helper["request_state"] = config.RequestState
		helper["resource_type"] = config.ResourceType
		helper["cluster"] = clusterCountMap[config.ComponentName]
		helper["ip_address"] = config.IPAddress
		rConfigs = append(rConfigs, helper)
	}
	return rConfigs
}

func parseDataMap(resourceData map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	resourcePropertyMapper := utils.ResourceMapper()
	for key, value := range resourceData {

		if i, ok := resourcePropertyMapper[key]; ok {
			key = i
		}
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice:
			parseArray(key, m, value.([]interface{}))
		case reflect.Map:
			parseMap(key, m, value.(map[string]interface{}))
		default:
			m[key] = convToString(value)
		}
	}
	return m
}

func parseMap(prefix string, m map[string]interface{}, data map[string]interface{}) {

	for key, value := range data {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Slice:
			parseArray(prefix+"."+key, m, value.([]interface{}))
		case reflect.Map:
			parseMap(prefix+"."+key, m, value.(map[string]interface{}))
		default:
			m[prefix+"."+key] = convToString(value)
		}
	}
}

func parseArray(prefix string, m map[string]interface{}, value []interface{}) {

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
					parseMap(prefix+"."+convToString(index), m, v.(map[string]interface{}))
				}
			}
		default:
			m[prefix+"."+convToString(index)] = convToString(val)
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
