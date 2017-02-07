package resources

import (
	"fmt"

	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"

	"github.com/errordeveloper/kubegen/pkg/resources/multi"
	"github.com/errordeveloper/kubegen/pkg/util"
)

func NewResourceGroupFromFile(path string) (*ResourceGroup, error) {
	const errfmt = "kubegen/resources: error reading resource group definition file %q – %v"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(errfmt, path, err)
	}

	var group *ResourceGroup
	if strings.HasSuffix(path, "yaml.kg") || strings.HasSuffix(path, ".yml.kg") {
		group, err = NewResourceGroupFromYAML(data)
		if err != nil {
			return nil, err
		}
		return group, nil
	}
	if strings.HasSuffix(path, ".kg") || strings.HasSuffix(path, ".json.kg") {
		group, err = NewResourceGroupFromHCL(data)
		if err != nil {
			return nil, err
		}
		return group, nil
	}

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
	multiGroup := &multi.ResourceGroup{}
	group := &ResourceGroup{}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("kubegen/resources: error converting YAML to internal HCL representation – %v", err)
	}

	if err := util.NewFromHCL(multiGroup, jsonData); err != nil {
		return nil, err
	}

	// This copier works better here, deepcopier doesn't work for some reason
	if err := copier.Copy(&group, *multiGroup); err != nil {
		return nil, fmt.Errorf("kubegen/resources: error converting mutil.ResourceGroup to ResourceGroup – %v", err)
	}

	return group, nil
}
