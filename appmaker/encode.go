package appmaker

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
	"k8s.io/client-go/pkg/api/v1"
	_ "k8s.io/client-go/pkg/apis/extensions/install"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func makeCodec(contentType string, pretty bool) (runtime.Codec, error) {
	serializerInfo, ok := runtime.SerializerInfoForMediaType(
		api.Codecs.SupportedMediaTypes(),
		contentType,
	)

	if !ok {
		return nil, fmt.Errorf("Unable to create a serializer")
	}

	serializer := serializerInfo.Serializer

	if pretty && serializerInfo.PrettySerializer != nil {
		serializer = serializerInfo.PrettySerializer
	}

	codec := api.Codecs.CodecForVersions(
		serializer,
		serializer,
		schema.GroupVersions(
			[]schema.GroupVersion{
				v1.SchemeGroupVersion,
				v1beta1.SchemeGroupVersion,
			},
		),
		runtime.InternalGroupVersioner,
	)

	return codec, nil
}

func (i *App) encodeList(contentType string, pretty bool) ([]byte, error) {
	components := i.MakeList()

	codec, err := makeCodec(contentType, pretty)
	if err != nil {
		return nil, err
	}
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

func (i *App) EncodeListToYAML() ([]byte, error) {
	return i.encodeList("application/yaml", false)
}

func (i *App) EncodeListToJSON() ([]byte, error) {
	return i.encodeList("application/json", false)
}

func (i *App) EncodeListToPrettyJSON() ([]byte, error) {
	return i.encodeList("application/json", true)
}
