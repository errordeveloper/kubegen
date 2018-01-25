package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ulule/deepcopier"
)

func (i ReplicaSet) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *ReplicaSet) Convert(localGroup *Group) (*appsv1.ReplicaSet, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	pod, err := MakePod(meta, i.Pod)
	if err != nil {
		return nil, fmt.Errorf("unable to define pod for ReplicaSet %q – %v", i.Name, err)
	}

	replicaSetSpec := appsv1.ReplicaSetSpec{
		Template: *pod,
		Replicas: &i.Replicas,
	}

	deepcopier.Copy(i).To(&replicaSetSpec)

	if len(i.Selector) == 0 {
		replicaSetSpec.Selector = &metav1.LabelSelector{MatchLabels: meta.Labels}
	} else {
		replicaSetSpec.Selector = &metav1.LabelSelector{MatchLabels: i.Selector}
	}

	replicaSet := appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta,
		Spec:       replicaSetSpec,
	}

	return &replicaSet, nil
}
