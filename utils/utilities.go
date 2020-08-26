package utils

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/op/go-logging"
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

// UnmarshalJSONStringIfNecessary parses value and if it's JSON string, unmarshal it
func UnmarshalJSONStringIfNecessary(field string, value interface{}) interface{} {
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
