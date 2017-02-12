package resources

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/ulule/deepcopier"
)

func (i DaemonSet) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *DaemonSet) Convert(localGroup *Group) (*v1beta1.DaemonSet, error) {
	meta := i.Metadata.Convert(i.Name)

	pod, err := MakePod(meta, i.Pod)
	if err != nil {
		return nil, fmt.Errorf("unable to define pod for DaemonSet  %q – %v", i.Name, err)
	}

	daemonSetSpec := v1beta1.DaemonSetSpec{
		Template: *pod,
	}

	deepcopier.Copy(i).To(&daemonSetSpec)

	if len(i.Selector) == 0 {
		daemonSetSpec.Selector = &metav1.LabelSelector{MatchLabels: meta.Labels}
	} else {
		daemonSetSpec.Selector = &metav1.LabelSelector{MatchLabels: i.Selector}
	}

	daemonSet := v1beta1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta,
		Spec:       daemonSetSpec,
	}

	return &daemonSet, nil
}
