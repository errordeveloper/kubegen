package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/ulule/deepcopier"
)

func (i ReplicaSet) ToObject() runtime.Object {
	return runtime.Object(i.Convert())
}

func (i *ReplicaSet) Convert() *v1beta1.ReplicaSet {
	meta := i.Metadata.Convert(i.Name)

	pod := MakePod(meta, i.Pod)

	replicaSetSpec := v1beta1.ReplicaSetSpec{
		Template: *pod,
		Replicas: &i.Replicas,
	}

	deepcopier.Copy(i).To(&replicaSetSpec)

	if len(i.Selector) == 0 {
		replicaSetSpec.Selector = &metav1.LabelSelector{MatchLabels: meta.Labels}
	} else {
		replicaSetSpec.Selector = &metav1.LabelSelector{MatchLabels: i.Selector}
	}

	replicaSet := v1beta1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta,
		Spec:       replicaSetSpec,
	}

	return &replicaSet
}
