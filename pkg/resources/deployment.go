package resources

import (
	_ "fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/ulule/deepcopier"
)

func (i *Deployment) Convert() *v1beta1.Deployment {
	meta := i.Metadata.Convert(i.Name)

	pod := MakePod(meta, i.Pod)

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

	return &deployment
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
		} // TODO should probably erorr here
	}

	return deploymentStrategy
}
