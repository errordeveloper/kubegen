package multi

import (
	"k8s.io/client-go/pkg/api/v1"
)

type ResourceGroup struct {
	Namespace    string        `hcl:"Namespace"`
	Services     []Service     `hcl:"Services"`
	Deployments  []Deployment  `hcl:"Deployments"`
	ReplicaSets  []ReplicaSet  `hcl:"ReplicaSets"`
	DaemonSets   []DaemonSet   `hcl:"DaemonSets"`
	StatefulSets []StatefulSet `hcl:"StatefulSets"`
	ConfigMaps   []ConfigMap   `hcl:"ConfigMaps"`
}

type Metadata struct {
	Labels      map[string]string `hcl:"labels"`
	Annotations map[string]string `hcl:"annotations"`
}

type Deployment struct {
	Name                    string            `hcl:",key"`
	Replicas                int32             `hcl:"replicas"`
	Selector                map[string]string `hcl:"selector"`
	Metadata                `hcl:",squash"`
	Pod                     `hcl:",squash"`
	Strategy                DeploymentStrategy `hcl:"strategy"`
	MinReadySeconds         int32              `hcl:"minReadySeconds"`
	RevisionHistoryLimit    *int32             `hcl:"revisionHistoryLimit"`
	Paused                  bool               `hcl:"paused"`
	ProgressDeadlineSeconds *int32             `hcl:"progressDeadlineSeconds"`
}

type ReplicaSet struct {
	Name            string            `hcl:",key"`
	Replicas        int32             `hcl:"replicas"`
	Selector        map[string]string `hcl:"selector"`
	Metadata        `hcl:",squash"`
	Pod             `hcl:",squash"`
	MinReadySeconds int32 `hcl:"minReadySeconds"`
}

type DaemonSet struct {
	Name     string            `hcl:",key"`
	Selector map[string]string `hcl:"selector"`
	Metadata `hcl:",squash"`
	Pod      `hcl:",squash"`
}

type StatefulSet struct {
	Name                 string            `hcl:",key"`
	Replicas             int32             `hcl:"replicas"`
	Selector             map[string]string `hcl:"selector"`
	Metadata             `hcl:",squash"`
	Pod                  `hcl:",squash"`
	VolumeClaimTemplates []v1.PersistentVolumeClaim `hcl:"volumeClaimTemplates"`
	ServiceName          string                     `hcl:"serviceName"`
}

type ConfigMap struct {
	Name          string `hcl:",key"`
	Metadata      `hcl:",squash"`
	Data          map[string]string      `hcl:"data"`
	DataFromFiles []string               `hcl:"dataFromFiles"`
	DataToJSON    map[string]interface{} `hcl:"dataToJSON"`
	DataToYAML    map[string]interface{} `hcl:"dataToYAML"`
}

type Pod struct {
	Annotations                   map[string]string   `hcl:"podAnnotations"`
	Volumes                       []Volume            `hcl:"volumes"`
	InitContainers                []Container         `hcl:"initContainers"`
	Containers                    []Container         `hcl:"containers"`
	RestartPolicy                 v1.RestartPolicy    `hcl:"restartPolicy"`
	TerminationGracePeriodSeconds *int64              `hcl:"terminationGracePeriodSeconds"`
	ActiveDeadlineSeconds         *int64              `hcl:"activeDeadlineSeconds"`
	DNSPolicy                     v1.DNSPolicy        `hcl:"dnsPolicy"`
	NodeSelector                  map[string]string   `hcl:"nodeSelector"`
	ServiceAccountName            string              `hcl:"serviceAccountName"`
	NodeName                      string              `hcl:"nodeName"`
	HostNetwork                   bool                `hcl:"hostNetwork"`
	HostPID                       bool                `hcl:"hostPID"`
	HostIPC                       bool                `hcl:"hostIPC"`
	SecurityContext               *PodSecurityContext `hcl:"securityContext"`
	ImagePullSecrets              []string            `hcl:"imagePullSecrets"`
	Hostname                      string              `hcl:"hostname"`
	Subdomain                     string              `hcl:"subdomain"`
	Affinity                      *v1.Affinity        `hcl:"affinity"`
	SchedulerName                 string              `json:"schedulerName"`
}

type Container struct {
	Name       string          `hcl:",key"`
	Image      string          `hcl:"image"`
	Command    []string        `hcl:"command"`
	Args       []string        `hcl:"args"`
	WorkingDir string          `hcl:"workDir"`
	Ports      []ContainerPort `hcl:"ports"`
	// EnvFrom []EnvFromSource
	Env                      map[string]string           `hcl:"env"`
	Resources                ResourceRequirements        `hcl:"resources"`
	Mounts                   []Mount                     `hcl:"volumeMounts"`
	LivenessProbe            *Probe                      `hcl:"livenessProbe"`
	ReadinessProbe           *Probe                      `hcl:"readinessProbe"`
	Lifecycle                *Lifecycle                  `hcl:"lifecycle"`
	TerminationMessagePath   string                      `hcl:"terminationMessagePath"`
	TerminationMessagePolicy v1.TerminationMessagePolicy `hcl:"terminationMessagePolicy"`
	ImagePullPolicy          v1.PullPolicy               `hcl:"imagePullPolicy"`
	SecurityContext          *SecurityContext            `hcl:"securityContext"`
	Stdin                    bool                        `hcl:"stdin"`
	StdinOnce                bool                        `hcl:"stdinOnce"`
	TTY                      bool                        `hcl:"tty"`
}

type ContainerPort struct {
	Name          string      `hcl:",key"`
	HostPort      int32       `hcl:"hostPort"`
	ContainerPort int32       `hcl:"containerPort"`
	Protocol      v1.Protocol `hcl:"protocol"`
	HostIP        string      `hcl:"hostIP"`
}

type Volume struct {
	Name         string `hcl:",key"`
	VolumeSource `hcl:",squash"`
}

// TODO: Figure out how to generate or import these
type VolumeSource struct {
	HostPath *HostPathVolumeSource `hcl:"hostPath"`
	EmptyDir *EmptyDirVolumeSource `hcl:"emptyDir"`
	Secret   *SecretVolumeSource   `hcl:"secret"`
}

type HostPathVolumeSource struct {
	Path string `hcl:"path"`
}

type EmptyDirVolumeSource struct {
	Medium v1.StorageMedium `hcl:"medium"`
}

type SecretVolumeSource struct {
	SecretName  string      `hcl:"secretName"`
	Items       []KeyToPath `hcl:"item"`
	DefaultMode int32       `hcl:"defaultMode"`
	Optional    bool        `hcl:"optional"`
}

type KeyToPath struct {
	Key  string `hcl:",key"`
	Path string `hcl:"path"`
	Mode *int32 `hcl:"mode"`
}

type Mount struct {
	Name      string `hcl:",key"`
	ReadOnly  bool   `hcl:"readOnly"`
	MountPath string `hcl:"mountPath"`
	SubPath   string `hcl:subPath"`
}

type Probe struct {
	Handler             `hcl:",squash"`
	InitialDelaySeconds int32 `hcl:"initialDelaySeconds"`
	TimeoutSeconds      int32 `hcl:"timeoutSeconds"`
	PeriodSeconds       int32 `hcl:"periodSeconds"`
	SuccessThreshold    int32 `hcl:"successThreshold"`
	FailureThreshold    int32 `hcl:"failureThreshold"`
}

type Lifecycle struct {
	PostStart *Handler `json:"postStart"`
	PreStop   *Handler `json:"preStop"`
}

type Handler struct {
	Exec      *ExecAction      `hcl:"exec"`
	HTTPGet   *HTTPGetAction   `hcl:"httpGet"`
	TCPSocket *TCPSocketAction `hcl:"tcpSocket"`
}

type ExecAction struct {
	Command []string `hcl:"command"`
}

type HTTPGetAction struct {
	Path        string            `hcl:"path"`
	Port        int32             `hcl:"port"`
	PortName    string            `hcl:"portName"`
	Host        string            `hcl:"host"`
	Scheme      v1.URIScheme      `hcl:"scheme"`
	HTTPHeaders map[string]string `hcl:"headers"`
}

type TCPSocketAction struct {
	Port     int32  `hcl:"port"`
	PortName string `hcl:"portName"`
}

type SecurityContext struct {
	Capabilities           *Capabilities      `hcl:"capabilities"`
	Privileged             *bool              `hcl:"privileged"`
	SELinuxOptions         *v1.SELinuxOptions `hcl:"seLinuxOptions"`
	RunAsUser              *int64             `hcl:"runAsUser"`
	RunAsNonRoot           *bool              `hcl:"runAsNonRoot"`
	ReadOnlyRootFilesystem *bool              `hcl:"readOnlyRootFilesystem"`
}

type PodSecurityContext struct {
	SELinuxOptions     *v1.SELinuxOptions `hcl:"seLinuxOptions"`
	RunAsUser          *int64             `hcl:"runAsUser"`
	RunAsNonRoot       *bool              `hcl:"runAsNonRoot"`
	SupplementalGroups []int64            `hcl:"supplementalGroups"`
	FSGroup            *int64             `hcl:"fsGroup"`
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
	Name                     string `hcl:",key"`
	Metadata                 `hcl:",squash"`
	Ports                    []ServicePort      `hcl:"port"`
	Selector                 map[string]string  `hcl:"selector"`
	ClusterIP                string             `hcl:"clusterIP"`
	Type                     v1.ServiceType     `hcl:"type"`
	ExternalIPs              []string           `hcl:"externalIPs"`
	SessionAffinity          v1.ServiceAffinity `hcl:"sessionAffinity"`
	LoadBalancerIP           string             `hcl:"loadBalancerIP"`
	LoadBalancerSourceRanges []string           `hcl:"loadBalancerSourceRanges"`
	ExternalName             string             `hcl:"externalName"`
}

type ServicePort struct {
	Name           string      `hcl:",key"`
	Port           int32       `hcl:"port"`
	Protocol       v1.Protocol `hcl:"protocol"`
	TargetPort     int32       `hcl:"targetPort"`
	TargetPortName string      `hcl:"targetPortName"`
	NodePort       int32       `hcl:"nodePort"`
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
	MaxUnavailable      string `hcl:"maxUnavailable"`
	MaxUnavailableCount *int   `hcl:"maxUnavailableCount"`
	MaxSurge            string `hcl:"maxSurge"`
	MaxSurgeCount       *int   `hcl:"maxSurgeCount"`
}
