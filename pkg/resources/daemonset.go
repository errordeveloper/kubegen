package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/ulule/deepcopier"
)

func (i *DaemonSet) Convert() *v1beta1.DaemonSet {
	meta := i.Metadata.Convert(i.Name)

	pod := MakePod(meta, i.Pod)

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

	return &daemonSet
}
