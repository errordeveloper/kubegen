package resources

import (
	"fmt"

	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ulule/deepcopier"
)

func (i ConfigMap) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *ConfigMap) Convert(localGroup *Group) (*corev1.ConfigMap, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	configMap := corev1.ConfigMap{
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
			return nil, fmt.Errorf("cannot read ConfigMap data from file %q – %v", v, err)
		}
		configMap.Data[v] = string(data)
	}

	return &configMap, nil
}
