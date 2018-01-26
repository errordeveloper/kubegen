package resources

import (
	"encoding/json"
	_ "fmt"
	"strings"

	"reflect"

	corev1 "k8s.io/api/core/v1"
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

func listAppendFlattened(components *metav1.List, component runtime.RawExtension) error {
	if component.Object != nil {
		if strings.HasSuffix(component.Object.GetObjectKind().GroupVersionKind().Kind, "List") {
			// must use corev1, as it panics on obj.(*metav1.List) with
			// an amusing error message saying that *v1.List is not *v1.List
			list := component.Object.(*corev1.List)
			for _, item := range (*list).Items {
				// we attempt to recurse here, but most likely
				// we will have to try decoding component.Raw
				if err := listAppendFlattened(components, item); err != nil {
					return err
				}
			}
			return nil
		}
		components.Items = append(components.Items, component)
		return nil
	}
	obj, err := util.Decode(component.Raw)
	if err != nil {
		return err
	}
	return listAppendFlattened(components, runtime.RawExtension{Object: obj})
}

// this version is a little simpler, but doesn't handle very raw nested lists
// keep it here, in case full recusion leads to problems due to decoding or whatnot
/*
func listAppendFlattenedSemiRecursive(components *metav1.List, component runtime.RawExtension) {
	if component.Object != nil {
		if strings.HasSuffix(component.Object.GetObjectKind().GroupVersionKind().Kind, "List") {
			// must use corev1, as it panics on obj.(*metav1.List) with
			// an amusing error message saying that *v1.List is not *v1.List
			list := component.Object.(*corev1.List)
			for _, item := range (*list).Items {
				listAppendFlattenedSemiRecursive(components, item)
			}
			return
		}
	}
	components.Items = append(components.Items, component)
}
*/

func (i *Group) appendToList(components *metav1.List, component Convertable) error {
	obj, err := component.ToObject(i)
	if err != nil {
		return err
	}

	return listAppendFlattened(components, runtime.RawExtension{Object: obj})
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
