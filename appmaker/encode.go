package appmaker

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
	"k8s.io/client-go/pkg/api/v1"
	_ "k8s.io/client-go/pkg/apis/extensions/install"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var codec runtime.Codec

func init() {
	serializer, ok := runtime.SerializerInfoForMediaType(
		api.Codecs.SupportedMediaTypes(),
		"application/yaml", // TODO both of JSON and YAML don't give us `kind` & `apiVersion`, why?
	)

	if !ok {
		panic("Unable to create a serializer")
	}

	codec = api.Codecs.CodecForVersions(
		serializer.Serializer,
		serializer.Serializer,
		schema.GroupVersions(
			[]schema.GroupVersion{
				v1.SchemeGroupVersion,
				v1beta1.SchemeGroupVersion,
			},
		),
		runtime.InternalGroupVersioner,
	)
}

func (i *App) Encode() ([]byte, error) {
	components := i.MakeList()

	// XXX: uncommenting this results in the following error:
	// json: error calling MarshalJSON for type runtime.RawExtension: invalid character 'a' looking for beginning of value
	//if err := runtime.EncodeList(codec, components.Items); err != nil {
	//	return nil, err
	//}

	data, err := runtime.Encode(codec, components)
	if err != nil {
		return nil, err
	}

	return data, nil
}
