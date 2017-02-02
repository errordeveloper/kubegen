package apps

import "github.com/errordeveloper/kubegen/pkg/util"

func NewFromHCL(data []byte) (*App, error) {
	app := &App{}

	if err := util.NewFromHCL(app, data); err != nil {
		return nil, err
	}

	return app, nil
}
