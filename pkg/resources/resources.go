package resources

import (
	"io/ioutil"
	_ "reflect"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"

	"github.com/errordeveloper/kubegen/pkg/util"
)

// This package is designed for direct HCL API for Kubernetes objects,
// it simplifies object structure slightly by flattening some deeply
// nested fields into top-level fields, yet it doesn't attempt to provide
// too abstract from native object kind. It can be used to compose objects
// from HCL files or as an intermediate representation for higher-level
// abstraction, such as the appmaker package.

type ResourceGroup struct {
	Deployments []Deployment `hcl:"deployment"`
	Services    []Service    `hcl:"service"`
	Namespace   string       `hcl:"namespace"`
}

type Deployment struct {
	Name           string      `hcl:",key"`
	Containers     []Container `hcl:"container"`
	Replicas       int32       `hcl:"replicas"`
	Volumes        []Volume    `hcl:"volume"`
	Metadata       `hcl:",squash"`
	PodAnnotations map[string]string `hcl:"pod_annotations"`
}

type Metadata struct {
	Labels      map[string]string `hcl:"labels"`
	Annotations map[string]string `hcl:"annotations"`
}

type Container struct {
	Name   string            `hcl:",key"`
	Image  string            `hcl:"image"`
	Ports  []ContainerPort   `hcl:"port"`
	Mounts []Mount           `hcl:"mount"`
	Env    map[string]string `hcl:"env"`
}

type ContainerPort struct {
	Name          string      `hcl:",key"`
	HostPort      int32       `hcl:"host_port"`
	ContainerPort int32       `hcl:"container_port"`
	Protocol      v1.Protocol `hcl:"protocol"`
	HostIP        string      `hcl:"host_ip"`
}

type Volume struct {
	Name         string `hcl:",key"`
	VolumeSource `hcl:",squash"`
}

// TODO: Figure out how to generate or import these
type VolumeSource struct {
	HostPath *HostPathVolumeSource `hcl:"host_path"`
	EmptyDir *EmptyDirVolumeSource `hcl:"empty_dir"`
	Secret   *SecretVolumeSource   `hcl:"secret"`
}

type HostPathVolumeSource struct {
	Path string `hcl:"path"`
}

type EmptyDirVolumeSource struct {
	Medium v1.StorageMedium `hcl:"medium"`
}

type SecretVolumeSource struct {
	SecretName  string      `hcl:"secret_name"`
	Items       []KeyToPath `hcl:"item"`
	DefaultMode int32       `hcl:"default_mode"`
	Optional    bool        `hcl:"optional"`
}

type KeyToPath struct {
	Key  string `hcl:",key"`
	Path string `hcl:"path"`
	Mode *int32 `hcl:"mode"`
}

type Mount struct {
	Name      string `hcl:",key"`
	ReadOnly  bool   `hcl:"read_only"`
	MountPath string `hcl:"mount_path"`
	SubPath   string `hcl:sub_path"`
}

type Service struct {
	Name     string `hcl:",key"`
	Metadata `hcl:",squash"`
	Ports    []ServicePort     `hcl:"port"`
	Selector map[string]string `hcl:"selector"`
}

type ServicePort struct {
	Name       string             `hcl:",key"`
	Protocol   v1.Protocol        `hcl:"protocol"`
	Port       int32              `hcl:"port"`
	TargetPort intstr.IntOrString `hcl:"target_port"`
	NodePort   int32              `hcl:"node_port"`
}

func NewResourceGroupFromPath(path string) (*ResourceGroup, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	group := &ResourceGroup{}

	if err := util.NewFromHCL(group, data); err != nil {
		return nil, err
	}

	return group, nil
}

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
