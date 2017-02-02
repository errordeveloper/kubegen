package resources

import (
	"fmt"
	"io/ioutil"
	_ "reflect"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
	"k8s.io/client-go/pkg/api/v1"
	_ "k8s.io/client-go/pkg/apis/extensions/install"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	_ "k8s.io/client-go/pkg/util/intstr"

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
	Medium string `hcl:"medium"`
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
	Mode int32  `hcl:"mode"`
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
	Name       string      `hcl:",key"`
	Protocol   v1.Protocol `hcl:"protocol"`
	Port       int32       `hcl:"port"`
	TargetPort int32       `hcl:"target_port"`
	NodePort   int32       `hcl:"node_port"`
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

func (i *Container) Convert() v1.Container {
	container := v1.Container{Name: i.Name, Image: i.Image}

	i.maybeAddEnvVars(&container)

	// you'd think the types could be simply converted,
	// but it turns out they won't because tags are different...
	// Fortunatelly, this has changed in Go1.8!
	//container.Ports = []v1.ContainerPort(i.Ports)
	for _, port := range i.Ports {
		container.Ports = append(container.Ports, v1.ContainerPort(port))
	}

	return container
}

func MakePod(labels, podAnnotations map[string]string, containers []Container) *v1.PodTemplateSpec {
	meta := metav1.ObjectMeta{
		Labels:      labels,
		Annotations: podAnnotations,
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{},
	}

	for _, container := range containers {
		podSpec.Containers = append(podSpec.Containers, container.Convert())
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

	pod := MakePod(i.Metadata.Labels, i.PodAnnotations, i.Containers)

	deploymentSpec := v1beta1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: meta.Labels},
		Template: *pod,
		Replicas: &i.Replicas,
	}

	deployment := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta,
		Spec:       deploymentSpec,
	}

	return deployment
}

func (i *ResourceGroup) EncodeListToPrettyJSON() ([]byte, error) {
	return i.encodeList("application/json", true)
}

func makeCodec(contentType string, pretty bool) (runtime.Codec, error) {
	serializerInfo, ok := runtime.SerializerInfoForMediaType(
		api.Codecs.SupportedMediaTypes(),
		contentType,
	)

	if !ok {
		return nil, fmt.Errorf("Unable to create a serializer")
	}

	serializer := serializerInfo.Serializer

	if pretty && serializerInfo.PrettySerializer != nil {
		serializer = serializerInfo.PrettySerializer
	}

	codec := api.Codecs.CodecForVersions(
		serializer,
		serializer,
		schema.GroupVersions(
			[]schema.GroupVersion{
				v1.SchemeGroupVersion,
				v1beta1.SchemeGroupVersion,
			},
		),
		runtime.InternalGroupVersioner,
	)

	return codec, nil
}

func (i *ResourceGroup) encodeList(contentType string, pretty bool) ([]byte, error) {
	components := &api.List{}
	for _, deployment := range i.Deployments {
		components.Items = append(components.Items, runtime.Object(deployment.Convert()))
	}

	codec, err := makeCodec(contentType, pretty)
	if err != nil {
		return nil, err
	}

	data, err := runtime.Encode(codec, components)
	if err != nil {
		return nil, err
	}

	return data, nil
}
