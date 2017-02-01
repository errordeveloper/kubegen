package resources

import (
	_ "fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

type Module struct {
	Deployments []Deployment `hcl:"deployment"`
	Services    []Service    `hcl:"service"`
}

type Deployment struct {
	Name       string      `hcl:",key"`
	Metadata   ObjectMeta  `hcl:"metadata"`
	Containers []Container `hcl:"container"`
	Replicas   int32       `hcl:"replicas"`
	Volumes    []Volume    `hcl:"volume"`
}

type ObjectMeta struct {
	Labels      map[string]string `hcl:"labels"`
	Annotations map[string]string `hcl:"annotations"`
}

type Container struct {
	Image  string            `hcl:"image"`
	Ports  []ContainerPort   `hcl:"port"`
	Mounts []Mount           `hcl:"mount"`
	Env    map[string]string `hcl:"env"`
}

type ContainerPort struct {
	Name          string `hcl:",key"`
	Protocol      string `hcl:"protocol"`
	ContainerPort int32  `hcl:"container_port"`
	HostPort      int32  `hcl:"host_port"`
	HostIP        string `hcl:"host_ip"`
}

type Volume struct {
	Name         string `hcl:",key"`
	VolumeSource `hcl:",squash"`
}

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
	Name     string            `hcl:",key"`
	Metadata ObjectMeta        `hcl:"metadata"`
	Ports    []ServicePort     `hcl:"port"`
	Selector map[string]string `hcl:"selector"`
}

type ServicePort struct {
	Name       string `hcl:",key"`
	Protocol   string `hcl:"protocol"`
	Port       int32  `hcl:"port"`
	TargetPort int32  `hcl:"target_port"`
	NodePort   int32  `hcl:"node_port"`
}

func NewModuleFromPath(path string) (*Module, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	module := &Module{}

	manifest, err := hcl.Parse(string(data))
	if err != nil {
		return nil, err
	}

	if err := hcl.DecodeObject(module, manifest); err != nil {
		return nil, err
	}

	return module, nil
}
