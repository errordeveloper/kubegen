package util

import (
	"fmt"

	"github.com/hashicorp/hcl"
)

func NewFromHCL(obj interface{}, data []byte) error {
	manifest, err := hcl.Parse(string(data))
	if err != nil {
		return fmt.Errorf("kubegen/util: error parsing HCL – %v", err)
	}

	if err := hcl.DecodeObject(obj, manifest); err != nil {
		return fmt.Errorf("kubegen/util: error constructing an object from HCL – %v", err)
	}

	return nil
}
