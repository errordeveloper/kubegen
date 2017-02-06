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
	Namespace    string        `hcl:"namespace"`
	Services     []Service     `hcl:"service"`
	Deployments  []Deployment  `hcl:"deployment"`
	ReplicaSets  []ReplicaSet  `hcl:"replicaset"`
	DaemonSets   []DaemonSet   `hcl:"daemonset"`
	StatefulSets []StatefulSet `hcl:"statefulset"`
}

type Metadata struct {
	Labels      map[string]string `hcl:"labels"`
	Annotations map[string]string `hcl:"annotations"`
}

type Deployment struct {
	Name                    string            `hcl:",key" deepcopier:"skip"`
	Replicas                int32             `hcl:"replicas" deepcopier:"skip"`
	Selector                map[string]string `hcl:"selector" deepcopier:"skip"`
	Metadata                `hcl:",squash" deepcopier:"skip"`
	Pod                     `hcl:",squash" deepcopier:"skip"`
	Strategy                DeploymentStrategy `hcl:"strategy" deepcopier:"skip"`
	MinReadySeconds         int32              `hcl:"min_ready_seconds"`
	RevisionHistoryLimit    *int32             `hcl:"revision_history_limit"`
	Paused                  bool               `hcl:"paused"`
	ProgressDeadlineSeconds *int32             `hcl:"progress_deadline_seconds"`
}

type ReplicaSet struct {
	Name            string            `hcl:",key" deepcopier:"skip"`
	Replicas        int32             `hcl:"replicas" deepcopier:"skip"`
	Selector        map[string]string `hcl:"selector" deepcopier:"skip"`
	Metadata        `hcl:",squash" deepcopier:"skip"`
	Pod             `hcl:",squash" deepcopier:"skip"`
	MinReadySeconds int32 `hcl:"min_ready_seconds"`
}

type DaemonSet struct {
	Name     string            `hcl:",key" deepcopier:"skip"`
	Selector map[string]string `hcl:"selector" deepcopier:"skip"`
	Metadata `hcl:",squash" deepcopier:"skip"`
	Pod      `hcl:",squash" deepcopier:"skip"`
}

type StatefulSet struct {
	Name                 string            `hcl:",key" deepcopier:"skip"`
	Replicas             int32             `hcl:"replicas" deepcopier:"skip"`
	Selector             map[string]string `hcl:"selector" deepcopier:"skip"`
	Metadata             `hcl:",squash" deepcopier:"skip"`
	Pod                  `hcl:",squash" deepcopier:"skip"`
	VolumeClaimTemplates []v1.PersistentVolumeClaim
	ServiceName          string
}

type Pod struct {
	Annotations                   map[string]string   `hcl:"pod_annotations" deepcopier:"skip"`
	Volumes                       []Volume            `hcl:"volume" deepcopier:"skip"`
	InitContainers                []Container         `hcl:"init_container" deepcopier:"skip"`
	Containers                    []Container         `hcl:"container" deepcopier:"skip"`
	RestartPolicy                 v1.RestartPolicy    `hcl:"restart_policy"`
	TerminationGracePeriodSeconds *int64              `hcl:"termination_grace_period_seconds"`
	ActiveDeadlineSeconds         *int64              `hcl:"active_deadline_seconds"`
	DNSPolicy                     v1.DNSPolicy        `hcl:"dns_policy"`
	NodeSelector                  map[string]string   `hcl:"node_selector"`
	ServiceAccountName            string              `hcl:"service_account_name"`
	NodeName                      string              `hcl:"node_name"`
	HostNetwork                   bool                `hcl:"host_network"`
	HostPID                       bool                `hcl:"host_pid"`
	HostIPC                       bool                `hcl:"host_ipc"`
	SecurityContext               *PodSecurityContext `hcl:"security_context" deepcopier:"skip"`
	ImagePullSecrets              []string            `hcl:"image_pull_secrets"`
	Hostname                      string              `hcl:"hostname"`
	Subdomain                     string              `hcl:"subdomain"`
	Affinity                      *v1.Affinity        `hcl:"affinity"`
	SchedulerName                 string              `json:"scheduler_name"`
}

type Container struct {
	Name       string          `hcl:",key" deepcopier:"skip"`
	Image      string          `hcl:"image" deepcopier:"skip"`
	Command    []string        `hcl:"command"`
	Args       []string        `hcl:"args"`
	WorkingDir string          `hcl:"work_dir"`
	Ports      []ContainerPort `hcl:"port" deepcopier:"skip"`
	// EnvFrom []EnvFromSource
	Env                      map[string]string           `hcl:"env" deepcopier:"skip"`
	Resources                ResourceRequirements        `hcl:"resources" deepcopier:"skip"`
	Mounts                   []Mount                     `hcl:"mount" deepcopier:"skip"`
	LivenessProbe            *Probe                      `hcl:"liveness_probe" deepcopier:"skip"`
	ReadinessProbe           *Probe                      `hcl:"readiness_probe" deepcopier:"skip"`
	Lifecycle                *Lifecycle                  `hcl:"lifecycle" deepcopier:"skip"`
	TerminationMessagePath   string                      `hcl:"termination_message_path"`
	TerminationMessagePolicy v1.TerminationMessagePolicy `hcl:"termination_message_policy"`
	ImagePullPolicy          v1.PullPolicy               `hcl:"image_pull_policy"`
	SecurityContext          *SecurityContext            `hcl:"security_context" deepcopier:"skip"`
	Stdin                    bool                        `hcl:"stdin"`
	StdinOnce                bool                        `hcl:"stdin_once"`
	TTY                      bool                        `hcl:"tty"`
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
	Handler             `hcl:",squash" deepcopier:"skip"`
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
	Port        int32             `hcl:"port" deepcopier:"skip"`
	PortName    string            `hcl:"port_name" deepcopier:"skip"`
	Host        string            `hcl:"host"`
	Scheme      v1.URIScheme      `hcl:"scheme"`
	HTTPHeaders map[string]string `hcl:"headers" deepcopier:"skip"`
}

type TCPSocketAction struct {
	Port     int32  `hcl:"port"`
	PortName string `hcl:"port_name"`
}

type SecurityContext struct {
	Capabilities           *Capabilities      `hcl:"capabilities"`
	Privileged             *bool              `hcl:"privileged"`
	SELinuxOptions         *v1.SELinuxOptions `hcl:"selinux_options"`
	RunAsUser              *int64             `hcl:"run_as_user"`
	RunAsNonRoot           *bool              `hcl:"run_as_non_root"`
	ReadOnlyRootFilesystem *bool              `hcl:"read_only_root_filesystem"`
}

type PodSecurityContext struct {
	SELinuxOptions     *v1.SELinuxOptions `hcl:"selinux_options"`
	RunAsUser          *int64             `hcl:"run_as_user"`
	RunAsNonRoot       *bool              `hcl:"run_as_non_root"`
	SupplementalGroups []int64            `hcl:"supplemental_groups"`
	FSGroup            *int64             `hcl:"fs_group"`
}

type SELinuxOptions struct {
	User  string `hcl:"user"`
	Role  string `hcl:"role"`
	Type  string `hcl:"type"`
	Level string `hcl:"level"`
}

type Capabilities struct {
	Add  []v1.Capability `json:"add"`
	Drop []v1.Capability `json:"drop"`
}

type Service struct {
	Name                     string `hcl:",key" deepcopier:"skip"`
	Metadata                 `hcl:",squash" deepcopier:"skip"`
	Ports                    []ServicePort      `hcl:"port" deepcopier:"skip"`
	Selector                 map[string]string  `hcl:"selector" deepcopier:"skip"`
	ClusterIP                string             `hcl:"cluster_ip"`
	Type                     v1.ServiceType     `hcl:"type"`
	ExternalIPs              []string           `hcl:"external_ips"`
	SessionAffinity          v1.ServiceAffinity `hcl:"session_affinity"`
	LoadBalancerIP           string             `hcl:"load_balancer_ip"`
	LoadBalancerSourceRanges []string           `hcl:"load_balancer_source_ranges"`
	ExternalName             string             `hcl:"external_name"`
}

type ServicePort struct {
	Name           string      `hcl:",key"`
	Port           int32       `hcl:"port"`
	Protocol       v1.Protocol `hcl:"protocol"`
	TargetPort     int32       `hcl:"target_port"`
	TargetPortName string      `hcl:"target_port_name"`
	NodePort       int32       `hcl:"node_port"`
}

type ResourceRequirements struct {
	Limits   map[string]string `json:"limits"`
	Requests map[string]string `json:"requests"`
}

type DeploymentStrategy struct {
	Type                    string `hcl:"type"`
	RollingUpdateDeployment `hcl:",squash"`
}

type RollingUpdateDeployment struct {
	MaxUnavailable      string `hcl:"max_unavailable"`
	MaxUnavailableCount *int   `hcl:"max_unavailable_count"`
	MaxSurge            string `hcl:"max_surge"`
	MaxSurgeCount       *int   `hcl:"max_surge_count"`
}
