package resources

import (
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"
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
