package vra7

import (
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/terraform-provider-vra7/sdk"
)

func resourceConfigurationSchema(optional bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: optional,
		Computed: !optional,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"component_name": {
					Type:     schema.TypeString,
					Optional: optional,
					Computed: !optional,
				},
				"configuration": {
					Type:     schema.TypeMap,
					Optional: optional,
					Computed: !optional,
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
					Removed:  "The ip_address is removed from here and available under resource_state",
				},
				"request_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"request_state": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"resource_state": {
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
							"state": {
								Type:     schema.TypeMap,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
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

		rConfig := sdk.ResourceConfigurationStruct{
			ComponentName: configMap["component_name"].(string),
			Configuration: configMap["configuration"].(map[string]interface{}),
			Cluster:       configMap["cluster"].(int),
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
		resourceStateList := make([]map[string]interface{}, 0)
		for _, resourceState := range config.ResourceState {
			rsMap := make(map[string]interface{})
			rsMap["resource_id"] = resourceState.ResourceID
			rsMap["resource_type"] = resourceState.ResourceType
			rsMap["status"] = resourceState.Status
			rsMap["name"] = resourceState.Name
			rsMap["date_created"] = resourceState.DateCreated
			rsMap["ip_address"] = resourceState.IPAddress
			stateMap, configurationMap := parseDataMap(resourceState.State, config.Configuration)
			rsMap["state"] = stateMap
			resourceStateList = append(resourceStateList, rsMap)
			helper["configuration"] = configurationMap
		}
		log.Critical("length of the rs list %v ", len(resourceStateList))
		helper["resource_state"] = resourceStateList
		helper["component_name"] = config.ComponentName
		helper["request_id"] = config.RequestID
		helper["parent_resource_id"] = config.ParentResourceID
		helper["request_state"] = config.RequestState
		helper["cluster"] = clusterCountMap[config.ComponentName]

		rConfigs = append(rConfigs, helper)
	}

	// j, _ := json.Marshal(rConfigs)
	// log.Critical("the struct in flatten is %v ", string(j))

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
