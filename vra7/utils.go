package vra7

import (
	"reflect"
	"strings"

	"github.com/vmware/terraform-provider-vra7/sdk"
	"github.com/vmware/terraform-provider-vra7/utils"
)

// UpdateResourceConfigurationMap updates the resource configuration with
//the deployment resource data if there is difference
// between the config data and deployment data, return true
func UpdateResourceConfigurationMap(
	resourceConfiguration map[string]interface{}, vmData map[string]map[string]interface{}) (map[string]interface{}, bool) {
	var changed bool
	for configKey1, configValue1 := range resourceConfiguration {
		for configKey2, configValue2 := range vmData {
			if strings.HasPrefix(configKey1, configKey2+".") {
				trimmedKey := strings.TrimPrefix(configKey1, configKey2+".")
				currentValue := configValue1
				updatedValue := utils.ConvertInterfaceToString(configValue2[trimmedKey])

				if updatedValue != "" && updatedValue != currentValue {
					resourceConfiguration[configKey1] = updatedValue
					changed = true
				}
			}
		}
	}
	return resourceConfiguration, changed
}

// ReplaceValueInRequestTemplate replaces the value for a given key in a catalog
// request template.
func ReplaceValueInRequestTemplate(templateInterface map[string]interface{}, field string, value interface{}) bool {
	var replaced bool
	//Iterate over the map to get field provided as an argument
	for key, val := range templateInterface {
		//If value type is map then set recursive call which will fiend field in one level down of map interface
		if reflect.ValueOf(val).Kind() == reflect.Map {
			replaced = ReplaceValueInRequestTemplate(val.(map[string]interface{}), field, value)
			if replaced {
				return true
			}
		} else if key == field && val != value {
			//If value type is not map then compare field name with provided field name
			//If both matches then update field value with provided value
			templateInterface[key] = value
			if reflect.ValueOf(value).Kind() == reflect.String {
				templateInterface[key] = utils.UnmarshalJSONStringIfNecessary(field, value)
			}
			return true
		}
	}
	return replaced
}

// AddValueToRequestTemplate modeled after replaceValueInRequestTemplate
// for values being added to template vs updating existing ones
func AddValueToRequestTemplate(templateInterface map[string]interface{}, field string, value interface{}) map[string]interface{} {
	//simplest case is adding a simple value. Leaving as a func in case there's a need to do more complicated additions later
	//	templateInterface[data]
	for k, v := range templateInterface {
		if reflect.ValueOf(v).Kind() == reflect.Map && k == "data" {
			template, _ := v.(map[string]interface{})
			_ = AddValueToRequestTemplate(template, field, value)
		} else { //if i == "data" {
			templateInterface[field] = utils.UnmarshalJSONStringIfNecessary(field, value)
		}
	}
	//Return updated map interface type
	return templateInterface
}

// ResourceMapper returns the mapping of resource attributes from ResourceView APIs
// to Catalog Item Request Template APIs
func ResourceMapper() map[string]string {
	m := make(map[string]string)
	m["MachineName"] = "name"
	m["MachineDescription"] = "description"
	m["MachineMemory"] = "memory"
	m["MachineStorage"] = "storage"
	m["MachineCPU"] = "cpu"
	m["MachineStatus"] = "status"
	m["MachineType"] = "type"
	return m
}

// GetConfiguration returns the configuration property for the componentName from the resource_configuration provided in the .tf file
func GetConfiguration(componentName string, resourceConfiguration []sdk.ResourceConfigurationStruct) map[string]interface{} {
	m := make(map[string]interface{})
	for _, elem := range resourceConfiguration {
		if elem.ComponentName == componentName {
			m = elem.Configuration
		}
	}
	return m
}
