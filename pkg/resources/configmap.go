package resources

import (
	"fmt"

	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/ulule/deepcopier"
)

func (i ConfigMap) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *ConfigMap) Convert(localGroup *Group) (*v1.ConfigMap, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	configMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
	}

	deepcopier.Copy(i).To(&configMap)

	if len(configMap.Data) == 0 {
		configMap.Data = make(map[string]string)
	}

	for _, v := range i.ReadFromFiles {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, fmt.Errorf("cannot read ConfigMap %q data from file %q – %v", v, err)
		}
		configMap.Data[v] = string(data)
	}

	return &configMap, nil
}
