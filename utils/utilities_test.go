package utils

import (
	"testing"
)

func TestFlatten(t *testing.T) {
	inside := make(map[string]interface{})
	inside["outside"] = "valid"
	outside := make(map[string]interface{})
	outside["test"] = inside
	actual := Flatten(outside)

	expected := make(map[string]interface{})
	expected["test.outside"] = "valid"

	for k := range actual {
		if actual[k] != expected[k] {
			t.Fatalf("Expected %s, got %s at key %s", expected, actual, k)
		}
	}
}

func TestFlattenComplex(t *testing.T) {
	deep := make(map[string]interface{})
	deep["outside"] = "valid"

	inside := make(map[string]interface{})
	inside["outside"] = "valid"
	inside["deep"] = deep

	outside := make(map[string]interface{})
	outside["test"] = inside
	actual := Flatten(outside)

	expected := make(map[string]interface{})
	expected["test.outside"] = "valid"
	expected["test.deep.outside"] = "valid"

	for k := range actual {
		if actual[k] != expected[k] {
			t.Fatalf("Expected %s, got %s at key %s", expected, actual, k)
		}
	}
}
