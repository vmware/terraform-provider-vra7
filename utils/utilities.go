package utils

import (
	"bytes"
	"encoding/json"
	"github.com/op/go-logging"
	"reflect"
	"strconv"
	"strings"
)

// terraform provider constants
const (
	// utility constants

	LoggerID = "terraform-provider-vra7"
)

var (
	log = logging.MustGetLogger(LoggerID)
)

// UnmarshalJSON  decodes json
func UnmarshalJSON(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}

// MarshalToJSON the object to JSON and convert to *bytes.Buffer
func MarshalToJSON(v interface{}) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(v)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// ConvertInterfaceToString cpnverts interface to string
func ConvertInterfaceToString(interfaceData interface{}) string {
	var stringData string
	if reflect.ValueOf(interfaceData).Kind() == reflect.Float64 {
		stringData =
			strconv.FormatFloat(interfaceData.(float64), 'f', 0, 64)
	} else if reflect.ValueOf(interfaceData).Kind() == reflect.Float32 {
		stringData =
			strconv.FormatFloat(interfaceData.(float64), 'f', 0, 32)
	} else if reflect.ValueOf(interfaceData).Kind() == reflect.Int {
		stringData = strconv.Itoa(interfaceData.(int))
	} else if reflect.ValueOf(interfaceData).Kind() == reflect.String {
		stringData = interfaceData.(string)
	} else if reflect.ValueOf(interfaceData).Kind() == reflect.Bool {
		stringData = strconv.FormatBool(interfaceData.(bool))
	}
	return stringData
}

// Parse value and if it's JSON string, unmarshal it
func UnmarshalJsonStringIfNecessary(field string, value interface{}) interface{} {
	// Cast value to string. Provider schema requires DeploymentConfiguration to be map[string]string
	stringValue, ok := value.(string)

	if !ok {
		log.Warning("Value of field=%v is not a string. Actual value %+v", field, value)
		return value
	}

	var jsonValue interface{}
	err := UnmarshalJSON([]byte(stringValue), &jsonValue)
	if err != nil {
		log.Debug("Value of field=%v is not a valid JSON string. Actual value %+v", field, value)
		return value
	}

	return jsonValue
}

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
				updatedValue := ConvertInterfaceToString(configValue2[trimmedKey])

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
func ReplaceValueInRequestTemplate(templateInterface map[string]interface{}, field string, value interface{}) (map[string]interface{}, bool) {
	var replaced bool
	//Iterate over the map to get field provided as an argument
	for key := range templateInterface {
		//If value type is map then set recursive call which will fiend field in one level down of map interface
		if reflect.ValueOf(templateInterface[key]).Kind() == reflect.Map {
			template, _ := templateInterface[key].(map[string]interface{})
			templateInterface[key], replaced = ReplaceValueInRequestTemplate(template, field, value)
			if replaced == true {
				return templateInterface, true
			}
		} else if key == field {
			//If value type is not map then compare field name with provided field name
			//If both matches then update field value with provided value
			templateInterface[key] = value
			return templateInterface, true
		}
	}
	//Return updated map interface type
	return templateInterface, replaced
}

// AddValueToRequestTemplate modeled after replaceValueInRequestTemplate
// for values being added to template vs updating existing ones
func AddValueToRequestTemplate(templateInterface map[string]interface{}, field string, value interface{}) map[string]interface{} {
	//simplest case is adding a simple value. Leaving as a func in case there's a need to do more complicated additions later
	//	templateInterface[data]
	for k, v := range templateInterface {
		if reflect.ValueOf(v).Kind() == reflect.Map && k == "data" {
			template, _ := v.(map[string]interface{})
			v = AddValueToRequestTemplate(template, field, value)
		} else { //if i == "data" {
			templateInterface[field] = value
		}
	}
	//Return updated map interface type
	return templateInterface
}
