package util

import (
	"encoding/json"

	"github.com/ghodss/yaml"
)

func toNonEmptyMap(obj interface{}) (map[string]interface{}, bool) {
	if v, ok := obj.(map[string]interface{}); ok && len(v) != 0 {
		return v, ok
	}
	return nil, false
}

func getMap(obj map[string]interface{}, key string) (map[string]interface{}, bool) {
	if v, ok := obj[key]; ok {
		if v, ok := toNonEmptyMap(v); ok {
			return v, ok
		}
	}
	return nil, false
}

func rangeOverNonEmptyMapsInSlice(obj map[string]interface{}, key string, iter func(map[string]interface{})) {
	if v, ok := obj[key]; ok {
		if v, ok := v.([]interface{}); ok && len(v) != 0 {
			for _, x := range v {
				if x, ok := x.(map[string]interface{}); ok && len(x) != 0 {
					iter(x)
				}
			}
		}
	}
}

func deleteKeyIfValueIsNil(obj map[string]interface{}, key string) {
	if v, ok := obj[key]; ok {
		if v == nil {
			delete(obj, key)
		}
	}
}

func deleteSubKeyIfValueIsNil(obj map[string]interface{}, k0, k1 string) {
	if v, ok := getMap(obj, k0); ok {
		deleteKeyIfValueIsNil(v, k1)
	}
	deleteKeyIfValueIsEmptyMap(obj, k0)
}

func deleteKeyIfValueIsEmptyMap(obj map[string]interface{}, key string) {
	if v, ok := obj[key]; ok {
		if v, ok := v.(map[string]interface{}); ok && len(v) == 0 {
			delete(obj, key)
		}
	}
}

func deleteSubKeyIfValueIsEmptyMap(obj map[string]interface{}, k0, k1 string) {
	if v, ok := getMap(obj, k0); ok {
		deleteKeyIfValueIsEmptyMap(v, k1)
	}
	deleteKeyIfValueIsEmptyMap(obj, k0)
}

func cleanupInnerSpec(item map[string]interface{}) {
	deleteSubKeyIfValueIsNil(item, "metadata", "creationTimestamp")
	deleteSubKeyIfValueIsEmptyMap(item, "status", "loadBalancer")

	deleteSubKeyIfValueIsEmptyMap(item, "spec", "strategy")
	deleteSubKeyIfValueIsEmptyMap(item, "spec", "updateStrategy")

	if spec, ok := getMap(item, "spec"); ok {
		if template, ok := getMap(spec, "template"); ok {
			if spec, ok := getMap(template, "spec"); ok {
				rangeOverNonEmptyMapsInSlice(spec, "initContainers", func(container map[string]interface{}) {
					deleteKeyIfValueIsEmptyMap(container, "resources")
					deleteKeyIfValueIsEmptyMap(container, "securityContext")
				})
				rangeOverNonEmptyMapsInSlice(spec, "containers", func(container map[string]interface{}) {
					deleteKeyIfValueIsEmptyMap(container, "resources")
					deleteKeyIfValueIsEmptyMap(container, "securityContext")
				})
			}

			deleteSubKeyIfValueIsNil(template, "metadata", "creationTimestamp")
		}
	}
}

func doCleanup(obj map[string]interface{}) {
	cleanupInnerSpec(obj)
	rangeOverNonEmptyMapsInSlice(obj, "items", func(item map[string]interface{}) {
		if item, ok := toNonEmptyMap(item); ok {
			cleanupInnerSpec(item)
		}
	})
}

func cleanup(contentType string, input []byte, pretty bool) ([]byte, error) {
	obj := make(map[string]interface{})
	var (
		output []byte
		err    error
	)
	switch contentType {
	case "application/yaml":
		if err = yaml.Unmarshal(input, &obj); err != nil {
			return nil, err
		}

		doCleanup(obj)

		if output, err = yaml.Marshal(obj); err != nil {
			return nil, err
		}
		return output, nil
	case "application/json":
		if err = json.Unmarshal(input, &obj); err != nil {
			return nil, err
		}

		doCleanup(obj)

		if pretty {
			output, err = json.MarshalIndent(obj, "", "  ")
		} else {
			output, err = json.Marshal(obj)
		}
		if err != nil {
			return nil, err
		}
		return output, nil
	default:
		return input, nil
	}
}
