package resources

import (
	"k8s.io/client-go/pkg/api/v1"
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

type Metadata struct {
	Labels      map[string]string `hcl:"labels"`
	Annotations map[string]string `hcl:"annotations"`
}

type Deployment struct {
	Name     string `hcl:",key"`
	Replicas int32  `hcl:"replicas"`
	Metadata `hcl:",squash"`
	Pod      `hcl:",squash"`
}

type Pod struct {
	Annotations                   map[string]string      `hcl:"pod_annotations"`
	Volumes                       []Volume               `hcl:"volume"`
	InitContainers                []Container            `hcl:"init_container"`
	Containers                    []Container            `hcl:"container"`
	RestartPolicy                 v1.RestartPolicy       `hcl:"restart_policy"`
	TerminationGracePeriodSeconds *int64                 `hcl:"termination_grace_period_seconds"`
	ActiveDeadlineSeconds         *int64                 `hcl:"active_deadline_seconds"`
	DNSPolicy                     v1.DNSPolicy           `hcl:"dns_policy"`
	NodeSelector                  map[string]string      `hcl:"node_selector"`
	ServiceAccountName            string                 `hcl:"service_account_name"`
	NodeName                      string                 `hcl:"node_name"`
	HostNetwork                   bool                   `hcl:"host_network"`
	HostPID                       bool                   `hcl:"host_pid"`
	HostIPC                       bool                   `hcl:"host_ipc"`
	SecurityContext               *v1.PodSecurityContext `hcl:"security_context"`
	ImagePullSecrets              []string               `hcl:"image_pull_secrets"`
	Hostname                      string                 `hcl:"hostname"`
	Subdomain                     string                 `hcl:"subdomain"`
	Affinity                      *v1.Affinity           `hcl:"affinity"`
	SchedulerName                 string                 `json:"scheduler_name"`
}

type Container struct {
	Name       string          `hcl:",key"`
	Image      string          `hcl:"image"`
	Command    []string        `hcl:"command"`
	Args       []string        `hcl:"args"`
	WorkingDir string          `hcl:"work_dir"`
	Ports      []ContainerPort `hcl:"port"`
	// EnvFrom []EnvFromSource
	Env                      map[string]string       `hcl:"env"`
	Resources                v1.ResourceRequirements `hcl:"resources"`
	Mounts                   []Mount                 `hcl:"mount"`
	LivenessProbe            *Probe                  `hcl:"liveness_probe"`
	ReadinessProbe           *Probe                  `hcl:"readiness_probe"`
	Lifecycle                `hcl:",squash"`
	TerminationMessagePath   string
	TerminationMessagePolicy v1.TerminationMessagePolicy
	ImagePullPolicy          v1.PullPolicy
	SecurityContext          *v1.SecurityContext
	Stdin                    bool
	StdinOnce                bool
	TTY                      bool
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

type Probe struct {
	Handler             `hcl:",squash"`
	InitialDelaySeconds int32 `hcl:"initial_delay_seconds"`
	TimeoutSeconds      int32 `hcl:"timeout_seconds"`
	PeriodSeconds       int32 `hcl:"period_seconds"`
	SuccessThreshold    int32 `hcl:"success_threshold"`
	FailureThreshold    int32 `hcl:"failure_threshold"`
}

type Lifecycle struct {
	PostStart *Handler `json:"post_start"`
	PreStop   *Handler `json:"pre_stop"`
}

type Handler struct {
	Exec      *ExecAction      `hcl:"exec"`
	HTTPGet   *HTTPGetAction   `hcl:"http_get"`
	TCPSocket *TCPSocketAction `hcl:"tcp_socket"`
}

type ExecAction struct {
	Command []string `hcl:"command"`
}

type HTTPGetAction struct {
	Path        string            `hcl:"path"`
	Port        int32             `hcl:"port"`
	PortName    string            `hcl:"port_name"`
	Host        string            `hcl:"host"`
	Scheme      v1.URIScheme      `hcl:"scheme"`
	HTTPHeaders map[string]string `hcl:"headers"`
}

type TCPSocketAction struct {
	Port     int32  `hcl:"port"`
	PortName string `hcl:"port_name"`
}

type Service struct {
	Name     string `hcl:",key"`
	Metadata `hcl:",squash"`
	Ports    []ServicePort     `hcl:"port"`
	Selector map[string]string `hcl:"selector"`
}

type ServicePort struct {
	Name           string      `hcl:",key"`
	Port           int32       `hcl:"port"`
	Protocol       v1.Protocol `hcl:"protocol"`
	TargetPort     int32       `hcl:"target_port"`
	TargetPortName string      `hcl:"target_port_name"`
	NodePort       int32       `hcl:"node_port"`
}
