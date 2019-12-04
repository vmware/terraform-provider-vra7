package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJsonStringIfNecessaryFunction(t *testing.T) {
	assertLocal := assert.New(t)
	const fieldName = "fieldName"

	var notStringValue = []interface{}{"2", 1, map[string]interface{}{
		"key1": "string",
		"key2": 1,
	}}
	computedValue1 := UnmarshalJSONStringIfNecessary(fieldName, notStringValue)
	assertLocal.Equal(notStringValue, computedValue1)

	var notJSONStringValue = "some custom value"
	computedValue2 := UnmarshalJSONStringIfNecessary(fieldName, notJSONStringValue)
	assertLocal.Equal(notJSONStringValue, computedValue2)

	var jSONStringValue = "[\"bg1\", \"bg2\"]"
	var expectedJSONValue = []interface{}{"bg1", "bg2"}
	computedValue3 := UnmarshalJSONStringIfNecessary(fieldName, jSONStringValue)
	assertLocal.Equal(expectedJSONValue, computedValue3)
}
