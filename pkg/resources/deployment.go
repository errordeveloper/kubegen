package resources

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/ulule/deepcopier"
)

func (i Deployment) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *Deployment) Convert(localGroup *Group) (*v1beta1.Deployment, error) {
	meta := i.Metadata.Convert(i.Name, localGroup.Namespace)

	pod, err := MakePod(meta, i.Pod)
	if err != nil {
		return nil, fmt.Errorf("unable to define pod for Deployment %q – %v", i.Name, err)
	}

	deploymentSpec := v1beta1.DeploymentSpec{
		Template: *pod,
		Replicas: &i.Replicas,
	}

	deepcopier.Copy(i).To(&deploymentSpec)

	if len(i.Selector) == 0 {
		deploymentSpec.Selector = &metav1.LabelSelector{MatchLabels: meta.Labels}
	} else {
		deploymentSpec.Selector = &metav1.LabelSelector{MatchLabels: i.Selector}
	}

	deploymentSpec.Strategy = i.Strategy.Convert()

	deployment := v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta,
		Spec:       deploymentSpec,
	}

	return &deployment, nil
}

func (i *DeploymentStrategy) Convert() v1beta1.DeploymentStrategy {
	deploymentStrategy := v1beta1.DeploymentStrategy{}

	if i.Type != "" {
		deploymentStrategy.Type = v1beta1.DeploymentStrategyType(i.Type)
		if i.Type == "RollingUpdate" {
			deploymentStrategy.RollingUpdate = &v1beta1.RollingUpdateDeployment{}

			if i.RollingUpdateDeployment.MaxUnavailable != "" {
				v := intstr.FromString(i.RollingUpdateDeployment.MaxUnavailable)
				deploymentStrategy.RollingUpdate.MaxUnavailable = &v
			} else if i.RollingUpdateDeployment.MaxUnavailableCount != nil {
				v := intstr.FromInt(*i.RollingUpdateDeployment.MaxUnavailableCount)
				deploymentStrategy.RollingUpdate.MaxUnavailable = &v
			}

			if i.RollingUpdateDeployment.MaxSurge != "" {
				v := intstr.FromString(i.RollingUpdateDeployment.MaxSurge)
				deploymentStrategy.RollingUpdate.MaxSurge = &v
			} else if i.RollingUpdateDeployment.MaxSurgeCount != nil {
				v := intstr.FromInt(*i.RollingUpdateDeployment.MaxSurgeCount)
				deploymentStrategy.RollingUpdate.MaxSurge = &v
			}
		}
	}

	return deploymentStrategy
}
