package resources

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"

	"github.com/ulule/deepcopier"
)

func (i StatefulSet) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *StatefulSet) Convert(localGroup *Group) (*v1beta1.StatefulSet, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	pod, err := MakePod(meta, i.Pod)
	if err != nil {
		return nil, fmt.Errorf("unable to define pod for StatefulSet %q – %v", i.Name, err)
	}

	statefulSetSpec := v1beta1.StatefulSetSpec{
		Template: *pod,
		Replicas: &i.Replicas,
	}

	deepcopier.Copy(i).To(&statefulSetSpec)

	if len(i.Selector) == 0 {
		statefulSetSpec.Selector = &metav1.LabelSelector{MatchLabels: meta.Labels}
	} else {
		statefulSetSpec.Selector = &metav1.LabelSelector{MatchLabels: i.Selector}
	}

	for _, volumeClaim := range i.VolumeClaimTemplates {
		statefulSetSpec.VolumeClaimTemplates = append(statefulSetSpec.VolumeClaimTemplates, volumeClaim)
	}

	statefulSet := v1beta1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta,
		Spec:       statefulSetSpec,
	}

	return &statefulSet, nil
}
