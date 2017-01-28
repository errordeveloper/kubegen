package appmaker

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
	"k8s.io/client-go/pkg/api/v1"
	_ "k8s.io/client-go/pkg/apis/extensions/install"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func (i *App) Encode() ([]byte, error) {
	components := &api.List{}
	codec := api.Codecs.LegacyCodec(v1.SchemeGroupVersion, v1beta1.SchemeGroupVersion)
	for _, component := range i.MakeAll() {
		switch component.manifest.Kind {
		case Deployment:
			components.Items = append(components.Items, runtime.Object(component.deployment))
		}

		if component.service != nil {
			components.Items = append(components.Items, runtime.Object(component.service))
		}
	}

	if err := runtime.EncodeList(codec, components.Items); err != nil {
		return nil, err
	}

	data, err := runtime.Encode(codec, components)
	if err != nil {
		return nil, err
	}

	return data, nil
}
