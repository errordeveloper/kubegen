package util

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
		return nil, fmt.Errorf("unable to create a serializer")
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

func EncodeList(list *api.List, contentType string, pretty bool) ([]byte, error) {
	codec, err := makeCodec(contentType, pretty)
	if err != nil {
		return nil, fmt.Errorf("kubegen/util: error creating codec for %q – %v", contentType, err)
	}
	// XXX: uncommenting this results in the following error:
	// json: error calling MarshalJSON for type runtime.RawExtension: invalid character 'a' looking for beginning of value
	//if err := runtime.EncodeList(codec, list.Items); err != nil {
	//	return nil, err
	//}

	data, err := runtime.Encode(codec, list)
	if err != nil {
		return nil, fmt.Errorf("kubegen/util: error encoding list to %q – %v", contentType, err)
	}

	return data, nil
}
