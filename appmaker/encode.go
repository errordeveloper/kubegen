package appmaker

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
)

func (i *App) Encode() ([]byte, error) {
	components := &api.List{}
	for _, component := range i.MakeAll() {
		switch component.manifest.Kind {
		case Deployment:
			components.Items = append(components.Items, runtime.Object(component.deployment))
		}

		if component.service != nil {
			components.Items = append(components.Items, runtime.Object(component.service))
		}
	}

	if err := runtime.EncodeList(latest.Codec, components.Items); err != nil {
		return nil, err
	}

	data, err := runtime.Encode(latest.Codec, components)
	if err != nil {
		return nil, err
	}

	return data, nil
}
