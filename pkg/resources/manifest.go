package resources

import (
	"fmt"

	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/errordeveloper/kubegen/pkg/util"
)

func NewResourceGroupFromFile(path string) (*ResourceGroup, error) {
	const errfmt = "kubegen/resources: error reading resource group definition file %q – %v"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(errfmt, path, err)
	}

	var group *ResourceGroup
	if strings.HasSuffix(path, "kg.yaml") || strings.HasSuffix(path, ".kg.yml") {
		group, err = NewResourceGroupFromYAML(data)
		if err != nil {
			return nil, err
		}
		return group, nil
	}
	//if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
	// TODO allow for vanilla YAML in a module
	//}
	if strings.HasSuffix(path, ".kg") || strings.HasSuffix(path, ".kg.hcl") || strings.HasSuffix(path, ".kg.json") {
		group, err = NewResourceGroupFromHCL(data)
		if err != nil {
			return nil, err
		}
		return group, nil
	}
	//if strings.HasSuffix(path, ".json") {
	// TODO allow for vanilla JSON in a module
	//}

	return nil, fmt.Errorf(errfmt, path, "unknown file extention")
}

func NewResourceGroupFromHCL(data []byte) (*ResourceGroup, error) {
	group := &ResourceGroup{}

	if err := util.NewFromHCL(group, data); err != nil {
		return nil, err
	}

	return group, nil
}

func NewResourceGroupFromYAML(data []byte) (*ResourceGroup, error) {
	group := &ResourceGroup{}

	if err := yaml.Unmarshal(data, group); err != nil {
		return nil, fmt.Errorf("kubegen/resources: error converting YAML to internal representation – %v", err)
	}

	return group, nil
}
