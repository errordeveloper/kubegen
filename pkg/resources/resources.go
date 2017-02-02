package resources

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/errordeveloper/kubegen/pkg/util"
)

func (i *Container) maybeAddEnvVars(container *v1.Container) {
	if len(i.Env) == 0 {
		return
	}

	keys := []string{}
	for k, _ := range i.Env {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	env := []v1.EnvVar{}
	for _, j := range keys {
		for k, v := range i.Env {
			if k == j {
				env = append(env, v1.EnvVar{Name: k, Value: v})
			}
		}
	}
	container.Env = env
}

func (i *Container) Convert() *v1.Container {
	container := v1.Container{Name: i.Name, Image: i.Image}

	i.maybeAddEnvVars(&container)

	// you'd think the types could be simply converted,
	// but it turns out they won't because tags are different...
	// Fortunatelly, this has changed in Go1.8!
	//container.Ports = []v1.ContainerPort(i.Ports)
	for _, port := range i.Ports {
		container.Ports = append(container.Ports, v1.ContainerPort(port))
	}

	for _, volumeMount := range i.Mounts {
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount(volumeMount))
	}

	return &container
}

func MakePod(labels, podAnnotations map[string]string, containers []Container, volumes []Volume) *v1.PodTemplateSpec {
	meta := metav1.ObjectMeta{
		Labels:      labels,
		Annotations: podAnnotations,
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{},
		Volumes:    []v1.Volume{},
	}

	for _, container := range containers {
		podSpec.Containers = append(podSpec.Containers, *container.Convert())
	}

	for _, volume := range volumes {
		v := v1.Volume{Name: volume.Name}
		if volume.HostPath != nil {
			s := v1.HostPathVolumeSource(*volume.VolumeSource.HostPath)
			v.VolumeSource.HostPath = &s
		}
		if volume.EmptyDir != nil {
			s := v1.EmptyDirVolumeSource(*volume.VolumeSource.EmptyDir)
			v.VolumeSource.EmptyDir = &s
		}
		if volume.Secret != nil {
			s := v1.SecretVolumeSource{
				SecretName:  volume.VolumeSource.Secret.SecretName,
				DefaultMode: &volume.VolumeSource.Secret.DefaultMode,
				Optional:    &volume.VolumeSource.Secret.Optional,
			}
			for _, item := range volume.VolumeSource.Secret.Items {
				s.Items = append(s.Items, v1.KeyToPath(item))
			}
			v.VolumeSource.Secret = &s
		}
		podSpec.Volumes = append(podSpec.Volumes, v)
	}

	pod := v1.PodTemplateSpec{
		ObjectMeta: meta,
		Spec:       podSpec,
	}

	return &pod
}

func (i *Deployment) Convert() *v1beta1.Deployment {
	meta := metav1.ObjectMeta{
		Name:        i.Name,
		Labels:      i.Metadata.Labels,
		Annotations: i.Metadata.Annotations,
	}

	pod := MakePod(i.Metadata.Labels, i.PodAnnotations, i.Containers, i.Volumes)

	deploymentSpec := v1beta1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: meta.Labels},
		Template: *pod,
		Replicas: &i.Replicas,
	}

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

func (i *Service) Convert() *v1.Service {
	meta := metav1.ObjectMeta{
		Name:        i.Name,
		Labels:      i.Metadata.Labels,
		Annotations: i.Metadata.Annotations,
	}

	serviceSpec := v1.ServiceSpec{
		Selector: i.Selector,
		Ports:    []v1.ServicePort{},
	}

	for _, port := range i.Ports {
		serviceSpec.Ports = append(serviceSpec.Ports, v1.ServicePort(port))
	}

	service := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
		Spec:       serviceSpec,
	}

	return &service
}

func (i *ResourceGroup) EncodeListToPrettyJSON() ([]byte, error) {
	return util.EncodeList(i.MakeList(), "application/json", true)
}

func (i *ResourceGroup) MakeList() *api.List {
	components := &api.List{}
	for _, component := range i.Deployments {
		components.Items = append(components.Items, runtime.Object(component.Convert()))
	}
	for _, component := range i.Services {
		components.Items = append(components.Items, runtime.Object(component.Convert()))
	}
	return components
}
