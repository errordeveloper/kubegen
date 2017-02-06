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
	ToObject() runtime.Object
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

func (i *ResourceGroup) EncodeListToPrettyJSON() ([]byte, error) {
	return util.EncodeList(i.MakeList(), "application/json", true)
}

func appendToList(components *api.List, component Convertable) {
	components.Items = append(components.Items, component.ToObject())
}

func (i *ResourceGroup) MakeList() *api.List {
	components := &api.List{}
	for _, component := range i.Services {
		appendToList(components, component)
	}
	for _, component := range i.Deployments {
		appendToList(components, component)
	}
	for _, component := range i.ReplicaSets {
		appendToList(components, component)
	}
	for _, component := range i.DaemonSets {
		appendToList(components, component)
	}
	for _, component := range i.StatefulSets {
		appendToList(components, component)
	}
	return components
}
