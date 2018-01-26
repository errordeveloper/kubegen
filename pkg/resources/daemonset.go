package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ulule/deepcopier"
)

func (i DaemonSet) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *DaemonSet) Convert(localGroup *Group) (*appsv1.DaemonSet, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	pod, err := MakePod(meta, i.Pod)
	if err != nil {
		return nil, fmt.Errorf("unable to define pod for DaemonSet  %q – %v", i.Name, err)
	}

	daemonSetSpec := appsv1.DaemonSetSpec{
		Template: *pod,
	}

	deepcopier.Copy(i).To(&daemonSetSpec)

	if len(i.Selector) == 0 {
		daemonSetSpec.Selector = &metav1.LabelSelector{MatchLabels: meta.Labels}
	} else {
		daemonSetSpec.Selector = &metav1.LabelSelector{MatchLabels: i.Selector}
	}

	daemonSet := appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta,
		Spec:       daemonSetSpec,
	}

	return &daemonSet, nil
}
