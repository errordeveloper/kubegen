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

// TODO add JSON tags...
type Group struct {
	Namespace    string        `yaml:"Namespace hcl:"namespace"`
	Services     []Service     `yaml:"Services" hcl:"service"`
	Deployments  []Deployment  `yaml:"Deployments" hcl:"deployment"`
	ReplicaSets  []ReplicaSet  `yaml:"ReplicaSets" hcl:"replicaset"`
	DaemonSets   []DaemonSet   `yaml:"DaemonSets" hcl:"daemonset"`
	StatefulSets []StatefulSet `yaml:"StatefulSet" hcl:"statefulset"`
	ConfigMaps   []ConfigMap   `yaml:"ConfigMaps" hcl:"configmap"`
}

type Metadata struct {
	Labels      map[string]string `yaml:"labels,omitempty" hcl:"labels"`
	Annotations map[string]string `yaml:"annotations,omitempty" hcl:"annotations"`
	Namespace   string            `yaml:"namespace,omitempty" hcl:"namespace"`
}

type Service struct {
	Name                     string `yaml:"name" hcl:",key" deepcopier:"skip"`
	Metadata                 `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Selector                 map[string]string  `yaml:"selector,omitempty" hcl:"selector" deepcopier:"skip"`
	Type                     v1.ServiceType     `yaml:"type,omitempty" hcl:"type"`
	Ports                    []ServicePort      `yaml:"ports,omitempty" hcl:"port" deepcopier:"skip"`
	ClusterIP                string             `yaml:"clusterIP,omitempty" hcl:"cluster_ip"`
	ExternalIPs              []string           `yaml:"externalIPs,omitempty" hcl:"external_ips"`
	ExternalName             string             `yaml:"externalName,omitempty" hcl:"external_name"`
	SessionAffinity          v1.ServiceAffinity `yaml:"sessionAffinity,omitempty" hcl:"session_affinity"`
	LoadBalancerIP           string             `yaml:"loadBalancerIP,omitempty" hcl:"load_balancer_ip"`
	LoadBalancerSourceRanges []string           `yaml:"loadBalancerSourceRanges,omitempty" hcl:"load_balancer_source_ranges"`
}

type Deployment struct {
	Name                    string `yaml:"name,omitempty" hcl:",key" deepcopier:"skip"`
	Metadata                `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Selector                map[string]string `yaml:"selector,omitempty" hcl:"selector" deepcopier:"skip"`
	Replicas                int32             `yaml:"replicas,omitempty" hcl:"replicas" deepcopier:"skip"`
	Pod                     `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Strategy                DeploymentStrategy `yaml:"strategy,omitempty" hcl:"strategy" deepcopier:"skip"`
	MinReadySeconds         int32              `yaml:"minReadySeconds,omitempty" hcl:"min_ready_seconds"`
	RevisionHistoryLimit    *int32             `yaml:"RevisionHistoryLimit,omitempty" hcl:"revision_history_limit"`
	Paused                  bool               `yaml:"paused,omitempty" hcl:"paused"`
	ProgressDeadlineSeconds *int32             `yaml:"progressDeadlineSeconds,omitempty" hcl:"progress_deadline_seconds"`
}

type ReplicaSet struct {
	Name            string `yaml:"name" hcl:",key" deepcopier:"skip"`
	Metadata        `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Selector        map[string]string `yaml:"selector,omitempty" hcl:"selector" deepcopier:"skip"`
	Replicas        int32             `yaml:"replicas,omitempty" hcl:"replicas" deepcopier:"skip"`
	Pod             `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	MinReadySeconds int32 `yaml:"minReadySeconds,omitempty" hcl:"min_ready_seconds"`
}

type DaemonSet struct {
	Name     string `yaml:"name" hcl:",key" deepcopier:"skip"`
	Metadata `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Selector map[string]string `yaml:"selector,omitempty" hcl:"selector" deepcopier:"skip"`
	Pod      `yaml:",inline" hcl:",squash" deepcopier:"skip"`
}

type StatefulSet struct {
	Name                 string `yaml:"name" hcl:",key" deepcopier:"skip"`
	Metadata             `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Selector             map[string]string `yaml:"selector,omitempty" hcl:"selector" deepcopier:"skip"`
	Replicas             int32             `yaml:"replicas,omitempty" hcl:"replicas" deepcopier:"skip"`
	Pod                  `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	VolumeClaimTemplates []v1.PersistentVolumeClaim `yaml:"volumeClaimTemplates,omitempty" hcl:"volume_claim" deepcopier:"skip"`
	ServiceName          string                     `yaml:"serviceName,omitempty" hcl:"service_name"`
}

type ConfigMap struct {
	Name          string `yaml:"name" hcl:",key" deepcopier:"skip"`
	Metadata      `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	Data          map[string]string      `yaml:"data,omitempty" hcl:"data"`
	DataFromFiles []string               `yaml:"dataFromFiles,omitempty" hcl:"data_from_files" deepcopier:"skip"`
	DataToJSON    map[string]interface{} `yaml:"dataToJSON,omitempty" hcl:"data_to_json" deepcopier:"skip"`
	DataToYAML    map[string]interface{} `yaml:"dataToYAML,omitempty" hcl:"data_to_yaml" deepcopier:"skip"`
}

type Pod struct {
	Annotations                   map[string]string   `yaml:"podAnnotations,omitempty" hcl:"pod_annotations" deepcopier:"skip"`
	Volumes                       []Volume            `yaml:"volumes,omitempty" hcl:"volume" deepcopier:"skip"`
	InitContainers                []Container         `yaml:"initContainers,omitempty" hcl:"init_container" deepcopier:"skip"`
	Containers                    []Container         `yaml:"containers,omitempty" hcl:"container" deepcopier:"skip"`
	RestartPolicy                 v1.RestartPolicy    `yaml:"restartPolicy,omitempty" hcl:"restart_policy"`
	TerminationGracePeriodSeconds *int64              `yaml:"terminationGracePeriodSeconds,omitempty" hcl:"termination_grace_period_seconds"`
	ActiveDeadlineSeconds         *int64              `yaml:"activeDeadlineSeconds,omitempty" hcl:"active_deadline_seconds"`
	DNSPolicy                     v1.DNSPolicy        `yaml:"dnsPolicy,omitempty" hcl:"dns_policy"`
	NodeSelector                  map[string]string   `yaml:"nodeSelector,omitempty" hcl:"node_selector"`
	ServiceAccountName            string              `yaml:"serviceAccountName,omitempty" hcl:"service_account_name"`
	NodeName                      string              `yaml:"nodeName,omitempty" hcl:"node_name"`
	HostNetwork                   bool                `yaml:"hostNetwork,omitempty" hcl:"host_network"`
	HostPID                       bool                `yaml:"hostPID,omitempty" hcl:"host_pid"`
	HostIPC                       bool                `yaml:"hostIPC,omitempty" hcl:"host_ipc"`
	SecurityContext               *PodSecurityContext `yaml:"securityContext,omitempty" hcl:"security_context" deepcopier:"skip"`
	ImagePullSecrets              []string            `yaml:"imagePullSecrets,omitempty" hcl:"image_pull_secrets"`
	Hostname                      string              `yaml:"hostname,omitempty" hcl:"hostname"`
	Subdomain                     string              `yaml:"subdomain,omitempty" hcl:"subdomain"`
	Affinity                      *v1.Affinity        `yaml:"affinity,omitempty" hcl:"affinity"`
	SchedulerName                 string              `yaml:"schedulerName,omitempty" hcl:"scheduler_name"`
}

type Container struct {
	Name       string          `yaml:"name" hcl:",key" deepcopier:"skip"`
	Image      string          `yaml:"image" hcl:"image" deepcopier:"skip"`
	Command    []string        `yaml:"command,omitempty" hcl:"command"`
	Args       []string        `yaml:"args,omitempty" hcl:"args"`
	WorkingDir string          `yaml:"workDir,omitempty" hcl:"work_dir"`
	Ports      []ContainerPort `yaml:"ports,omitempty" hcl:"port" deepcopier:"skip"`
	// EnvFrom []EnvFromSource
	Env                      map[string]string           `yaml:"env,omitempty" hcl:"env" deepcopier:"skip"`
	Resources                ResourceRequirements        `yaml:"resources,omitempty" hcl:"resources" deepcopier:"skip"`
	Mounts                   []Mount                     `yaml:"volumeMounts,omitempty" hcl:"mount" deepcopier:"skip"`
	LivenessProbe            *Probe                      `yaml:"livenessProbe,omitempty" hcl:"liveness_probe" deepcopier:"skip"`
	ReadinessProbe           *Probe                      `yaml:"readinessProbe,omitempty" hcl:"readiness_probe" deepcopier:"skip"`
	Lifecycle                *Lifecycle                  `yaml:"lifecycle,omitempty" hcl:"lifecycle" deepcopier:"skip"`
	TerminationMessagePath   string                      `yaml:"terminationMessagePath,omitempty" hcl:"termination_message_path"`
	TerminationMessagePolicy v1.TerminationMessagePolicy `yaml:"terminationMessagePolicy,omitempty" hcl:"termination_message_policy"`
	ImagePullPolicy          v1.PullPolicy               `yaml:"imagePullPolicy,omitempty" hcl:"image_pull_policy"`
	SecurityContext          *SecurityContext            `yaml:"securityContext,omitempty" hcl:"security_context" deepcopier:"skip"`
	Stdin                    bool                        `yaml:"stdin,omitempty" hcl:"stdin"`
	StdinOnce                bool                        `yaml:"stdinOnce,omitempty" hcl:"stdin_once"`
	TTY                      bool                        `yaml:"tty,omitempty" hcl:"tty"`
}

type ContainerPort struct {
	Name          string      `yaml:"name" hcl:",key"`
	HostPort      int32       `yaml:"hostPort,omitempty" hcl:"host_port"`
	ContainerPort int32       `yaml:"containerPort,omitempty" hcl:"container_port"`
	Protocol      v1.Protocol `yaml:"protocol,omitempty" hcl:"protocol"`
	HostIP        string      `yaml:"hostIP,omitempty" hcl:"host_ip"`
}

type ServicePort struct {
	Name           string      `yaml:"name" hcl:",key"`
	Port           int32       `yaml:"port,omitempty" hcl:"port"`
	Protocol       v1.Protocol `yaml:"protocol,omitempty" hcl:"protocol"`
	TargetPort     int32       `yaml:"targetPort,omitempty" hcl:"target_port"`
	TargetPortName string      `yaml:"targetPortName,omitempty" hcl:"target_port_name"`
	NodePort       int32       `yaml:"nodePort,omitempty" hcl:"node_port"`
}

type Volume struct {
	Name         string `yaml:"name,omitempty" hcl:",key"`
	VolumeSource `yaml:",inline" hcl:",squash"`
}

// TODO: Figure out how to generate or import these
type VolumeSource struct {
	HostPath  *HostPathVolumeSource  `yaml:"hostPath,omitempty" hcl:"host_path"`
	EmptyDir  *EmptyDirVolumeSource  `yaml:"emptyDir,omitempty" hcl:"empty_dir"`
	Secret    *SecretVolumeSource    `yaml:"secret,omitempty" hcl:"secret"`
	ConfigMap *ConfigMapVolumeSource `yaml:"configMap" hcl:"configmap"`
}

type HostPathVolumeSource struct {
	Path string `yaml:"path,omitempty" hcl:"path"`
}

type EmptyDirVolumeSource struct {
	Medium v1.StorageMedium `yaml:"medium,omitempty" hcl:"medium"`
}

type SecretVolumeSource struct {
	SecretName  string      `yaml:"secretName,omitempty" hcl:"secret_name"`
	Items       []KeyToPath `yaml:"items,omitempty" hcl:"item"`
	DefaultMode *int32      `yaml:"defaultMode,omitempty" hcl:"default_mode"`
	Optional    *bool       `yaml:"optional,omitempty" hcl:"optional"`
}

type ConfigMapVolumeSource struct {
	LocalObjectReference `yaml:",inline" hcl:",squash"`
	Items                []KeyToPath `yaml:"items,omitempty" hcl:"items"`
	DefaultMode          *int32      `yaml:"defaultMode,omitempty" hcl:"default_mode"`
	Optional             *bool       `json:"optional,omitempty" hcl:"optional"`
}

type LocalObjectReference struct {
	Name string `yaml:"name,omitempty" hcl:"name"`
}

type KeyToPath struct {
	Key  string `yaml:"key" hcl:",key"`
	Path string `yaml:"path" hcl:"path"`
	Mode *int32 `yaml:"mode,omitempty" hcl:"mode"`
}

type Mount struct {
	Name      string `yaml:"name" hcl:",key"`
	ReadOnly  bool   `yaml:"readOnly,omitempty" hcl:"read_only"`
	MountPath string `yaml:"mountPath,omitempty" hcl:"mount_path"`
	SubPath   string `yaml:"subPath,omitempty" hcl:sub_path"`
}

type Probe struct {
	Handler             `yaml:",inline" hcl:",squash" deepcopier:"skip"`
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" hcl:"initial_delay_seconds"`
	TimeoutSeconds      int32 `yaml:"timeoutSeconds,omitempty" hcl:"timeout_seconds"`
	PeriodSeconds       int32 `yaml:"periodSeconds,omitempty" hcl:"period_seconds"`
	SuccessThreshold    int32 `yaml:"successThreshold,omitempty" hcl:"success_threshold"`
	FailureThreshold    int32 `yaml:"failureThreshold,omitempty" hcl:"failure_threshold"`
}

type Lifecycle struct {
	PostStart *Handler `yaml:"postStart,omitempty" hcl:"post_start"`
	PreStop   *Handler `yaml:"preStop,omitempty" hcl:"pre_stop"`
}

type Handler struct {
	Exec      *ExecAction      `yaml:"exec,omitempty" hcl:"exec"`
	HTTPGet   *HTTPGetAction   `yaml:"httpGet,omitempty" hcl:"http_get"`
	TCPSocket *TCPSocketAction `yaml:"tcpSocket,omitempty" hcl:"tcp_socket"`
}

type ExecAction struct {
	Command []string `yaml:"command,omitempty" hcl:"command"`
}

type HTTPGetAction struct {
	Path        string            `yaml:"path,omitempty" hcl:"path"`
	Port        int32             `yaml:"port,omitempty" hcl:"port" deepcopier:"skip"`
	PortName    string            `yaml:"portName,omitempty" hcl:"port_name" deepcopier:"skip"`
	Host        string            `yaml:"host,omitempty" hcl:"host"`
	Scheme      v1.URIScheme      `yaml:"scheme,omitempty" hcl:"scheme"`
	HTTPHeaders map[string]string `yaml:"headers,omitempty" hcl:"headers" deepcopier:"skip"`
}

type TCPSocketAction struct {
	Port     int32  `yaml:"port,omitempty" hcl:"port"`
	PortName string `yaml:"portName,omitempty" hcl:"port_name"`
}

type SecurityContext struct {
	Capabilities           *Capabilities      `yaml:"capabilities,omitempty" hcl:"capabilities"`
	Privileged             *bool              `yaml:"privileged,omitempty" hcl:"privileged"`
	SELinuxOptions         *v1.SELinuxOptions `yaml:"seLinuxOptions,omitempty" hcl:"selinux_options"`
	RunAsUser              *int64             `yaml:"runAsUser,omitempty" hcl:"run_as_user"`
	RunAsNonRoot           *bool              `yaml:"runAsNonRoot,omitempty" hcl:"run_as_non_root"`
	ReadOnlyRootFilesystem *bool              `yaml:"readOnlyRootFilesystem,omitempty" hcl:"read_only_root_filesystem"`
}

type PodSecurityContext struct {
	SELinuxOptions     *v1.SELinuxOptions `yaml:"seLinuxOptions,omitempty" hcl:"selinux_options"`
	RunAsUser          *int64             `yaml:"runAsUser,omitempty" hcl:"run_as_user"`
	RunAsNonRoot       *bool              `yaml:"runAsNonRoot,omitempty" hcl:"run_as_non_root"`
	SupplementalGroups []int64            `yaml:"supplementalGroups,omitempty" hcl:"supplemental_groups"`
	FSGroup            *int64             `yaml:"fsGroup,omitempty" hcl:"fs_group"`
}

type SELinuxOptions struct {
	User  string `yaml:"user,omitempty" hcl:"user"`
	Role  string `yaml:"role,omitempty" hcl:"role"`
	Type  string `yaml:"type,omitempty" hcl:"type"`
	Level string `yaml:"level,omitempty" hcl:"level"`
}

type Capabilities struct {
	Add  []v1.Capability `yaml:"add,omitempty" hcl:"add"`
	Drop []v1.Capability `yaml:"drop,omitempty" hcl:"drop"`
}

type ResourceRequirements struct {
	Limits   map[string]string `yaml:"limits,omitempty" hcl:"limits"`
	Requests map[string]string `yaml:"requests,omitempty" hcl:"requests"`
}

type DeploymentStrategy struct {
	Type                    string `yaml:"type,omitempty" hcl:"type"`
	RollingUpdateDeployment `yaml:",inline,omitempty" hcl:",squash"`
}

type RollingUpdateDeployment struct {
	MaxUnavailable      string `yaml:"maxUnavailable,omitempty" hcl:"max_unavailable"`
	MaxUnavailableCount *int   `yaml:"maxUnavailableCount,omitempty" hcl:"max_unavailable_count"`
	MaxSurge            string `yaml:"maxSurge,omitempty" hcl:"max_surge"`
	MaxSurgeCount       *int   `yaml:"maxSurgeCount,omitempty" hcl:"max_surge_count"`
}
