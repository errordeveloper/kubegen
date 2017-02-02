package resources

import (
	"io/ioutil"

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
