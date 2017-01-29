package appmaker

import (
	"github.com/hashicorp/hcl"
)

func NewFromHCL(data []byte) (*App, error) {
	app := &App{}

	manifest, err := hcl.Parse(string(data))
	if err != nil {
		return nil, err
	}

	if err := hcl.DecodeObject(&app, manifest); err != nil {
		return nil, err
	}

	return app, nil
}
