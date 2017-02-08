package resources

import (
	_ "fmt"

	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"

	"github.com/errordeveloper/kubegen/pkg/util"
)

type Convertable interface {
	ToObject() (runtime.Object, error)
}

func exclusiveNonNil(args ...interface{}) *int {
	count := 0
	index := 0
	for k, v := range args {
		if !reflect.ValueOf(v).IsNil() {
			count = count + 1
			index = k
		}
	}

	if count == 0 || count > 1 {
		return nil
	} else {
		return &index
	}
}

func (i *Metadata) Convert(name string) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        name,
		Labels:      i.Labels,
		Annotations: i.Annotations,
	}

	if len(meta.Labels) == 0 {
		meta.Labels = map[string]string{"name": name}
	}

	return meta
}

func (i *ResourceGroup) EncodeListToYAML() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return []byte{}, err
	}
	return util.EncodeList(list, "application/yaml", false)
}

func (i *ResourceGroup) EncodeListToJSON() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return []byte{}, err
	}
	return util.EncodeList(list, "application/json", false)
}

func (i *ResourceGroup) EncodeListToPrettyJSON() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return []byte{}, err
	}
	return util.EncodeList(list, "application/json", true)
}

func appendToList(components *api.List, component Convertable) error {
	obj, err := component.ToObject()
	if err != nil {
		return err
	}
	components.Items = append(components.Items, obj)
	return nil
}

func (i *ResourceGroup) MakeList() (*api.List, error) {
	components := &api.List{}
	for _, component := range i.Services {
		if err := appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.Deployments {
		if err := appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.ReplicaSets {
		if err := appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.DaemonSets {
		if err := appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.StatefulSets {
		if err := appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.ConfigMaps {
		if err := appendToList(components, component); err != nil {
			return nil, err
		}
	}
	return components, nil
}
