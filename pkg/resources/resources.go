package resources

import (
	"encoding/json"
	_ "fmt"

	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/errordeveloper/kubegen/pkg/util"
)

type Convertable interface {
	ToObject(*Group) (runtime.Object, error)
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
	}
	return &index
}

func (i *Metadata) Convert(name string, localGroup *Group) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        name,
		Labels:      i.Labels,
		Annotations: i.Annotations,
		Namespace:   i.Namespace,
	}

	if localGroup != nil {
		if meta.Namespace == "" && localGroup.Namespace != "" {
			meta.Namespace = localGroup.Namespace
		}
	}

	if len(meta.Labels) == 0 {
		meta.Labels = map[string]string{"name": name}
	}

	return meta
}

func (i AnyResource) ToObject(localGroup *Group) (runtime.Object, error) {
	jsonData, err := json.Marshal(i.Object)
	if err != nil {
		return runtime.Object(nil), err
	}
	return util.Decode(jsonData)
}
func (i *Group) EncodeListToYAML() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	return util.EncodeList(list, "application/yaml", false)
}

func (i *Group) EncodeListToJSON() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	return util.EncodeList(list, "application/json", false)
}

func (i *Group) EncodeListToPrettyJSON() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	return util.EncodeList(list, "application/json", true)
}

func (i *Group) appendToList(components *metav1.List, component Convertable) error {
	obj, err := component.ToObject(i)
	if err != nil {
		return err
	}
	components.Items = append(components.Items, runtime.RawExtension{Object: obj})
	return nil
}

func (i *Group) MakeList() (*metav1.List, error) {
	components := &metav1.List{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
	}
	for _, component := range i.Deployments {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.ReplicaSets {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.DaemonSets {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.StatefulSets {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.ConfigMaps {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.Services {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	for _, component := range i.AnyResources {
		if err := i.appendToList(components, component); err != nil {
			return nil, err
		}
	}
	return components, nil
}
