package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalJsonStringIfNecessaryFunction(t *testing.T) {
	assertLocal := assert.New(t)
	const fieldName = "fieldName"

	var notStringValue = []interface{}{"2", 1, map[string]interface{}{
		"key1": "string",
		"key2": 1,
	}}
	computedValue1 := UnmarshalJsonStringIfNecessary(fieldName, notStringValue)
	assertLocal.Equal(notStringValue, computedValue1)

	var notJsonStringValue = "some custom value"
	computedValue2 := UnmarshalJsonStringIfNecessary(fieldName, notJsonStringValue)
	assertLocal.Equal(notJsonStringValue, computedValue2)

	var jsonStringValue = "[\"bg1\", \"bg2\"]"
	var expectedJsonValue = []interface{}{"bg1", "bg2"}
	computedValue3 := UnmarshalJsonStringIfNecessary(fieldName, jsonStringValue)
	assertLocal.Equal(expectedJsonValue, computedValue3)
}
