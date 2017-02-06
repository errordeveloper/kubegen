package resources

import (
	"fmt"

	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"

	"github.com/errordeveloper/kubegen/pkg/resources/multi"
	"github.com/errordeveloper/kubegen/pkg/util"
)

func NewResourceGroupFromPath(path string) (*ResourceGroup, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	group := &ResourceGroup{}

	if err := util.NewFromHCL(group, data); err != nil {
		return nil, err
	}

	return group, nil
}

func NewResourceGroupFromYAML(path string) (*ResourceGroup, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	multiGroup := &multi.ResourceGroup{}

	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}

	if err := util.NewFromHCL(multiGroup, data); err != nil {
		return nil, err
	}

	group := &ResourceGroup{}

	copier.Copy(&group, *multiGroup)

	return group, nil
}
