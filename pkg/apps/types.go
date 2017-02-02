package apps

import (
	_ "fmt"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type App struct {
	GroupName               string                     `hcl:"group_name"`
	Components              []AppComponent             `hcl:"-"`
	ComponentsFromImages    []AppComponentFromImage    `hcl:"component_from_image"`
	Templates               []AppComponentTemplate     `hcl:"component_template"`
	ComponentsFromTemplates []AppComponentFromTemplate `hcl:"component_from_template"`
	CommonEnv               map[string]string          `hcl:"common_env"`
}

type AppComponent struct {
	Image     string            `hcl:"-"`
	Name      string            `json:",omitempty" hcl:"name,omitempty"`
	Port      int32             `json:",omitempty" hcl:"port,omitempty"`
	Replicas  *int32            `json:",omitempty" hcl:"replicas,omitempty"`
	Flavor    string            `json:",omitempty" hcl:"flavor,omitempty"`
	Opts      AppComponentOpts  `json:",omitempty" hcl:"opts,omitempty"`
	Env       map[string]string `json:",omitempty" hcl:"env,omitempty"`
	CommonEnv []string          `json:",omitempty" hcl:"common_env,omitempty"`
	// Deployment, DaemonSet, StatefullSet ...etc
	Kind int `json:",omitempty" hcl:"kind,omitempty"`
	// It's probably okay for now, but we'd eventually want to
	// inherit properties defined outside of the AppComponent struct,
	// that it anything we'd use setters and getters for, so we might
	// want to figure out intermediate struct or just write more
	// some tests to see how things would work without that...
	basedOn              *AppComponent        `json:"-" hcl:"-"`
	BasedOnNamedTemplate string               `json:",omitempty" hcl:"based_on,omitempty"`
	Customize            GeneralCustomizer    `json:"-" hcl:"-"`
	CustomizeCotainers   ContainersCustomizer `json:"-" hcl:"-"`
	CustomizePod         PodCustomizer        `json:"-" hcl:"-"`
	CustomizeService     ServiceCustomizer    `json:"-" hcl:"-"`
	CustomizePorts       PortsCustomizer      `json:"-" hcl:"-"`
}

type AppComponentFromImage struct {
	Image        string `hcl:",key"`
	AppComponent `hcl:",squash"`
}

type AppComponentTemplate struct {
	TemplateName string `json:",omitempty" hcl:",key"`
	Image        string `json:",omitempty" hcl:"image"`
	AppComponent `json:",inline" hcl:",squash"`
}

// AppComponentFromTemplate is the same as AppComponentTemplate, but it is an alias, because
// it makes the code easier to read
type AppComponentFromTemplate AppComponentTemplate

// Everything we want to controll per-app, but it's not exposed to HCL directly
type AppParams struct {
	Namespace              string
	DefaultReplicas        int32
	DefaultPort            int32
	StandardLivenessProbe  *v1.Probe
	StandardReadinessProbe *v1.Probe
	templates              map[string]AppComponent
	commonEnv              map[string]string
}

// AppComponentOpts hold highlevel fields which map to a non-trivial settings
// within inside the object, often affecting sub-fields within sub-fields,
// for more trivial things (like hostNetwork) we have custom setters
type AppComponentOpts struct {
	PrometheusPath   string `json:",omitempty"`
	PrometheusScrape bool   `json:",omitempty"`
	// WithoutPorts implies WithoutService and WithoutStandardProbes
	WithoutPorts                   bool   `json:",omitempty"`
	WithoutStandardProbes          bool   `json:",omitempty"`
	WithoutStandardSecurityContext bool   `json:",omitempty"`
	HealthPath                     string `json:",omitempty"`
	LivenessPath                   string `json:",omitempty"`
	// XXX we can add these here, but may be they belong elsewhere?
	//WithProbes interface{}
	//WithSecurityContext interface{}
	// WithoutService disables building of the service
	WithoutService bool `json:",omitempty"`
}

type (
	GeneralCustomizer func(
		*AppComponentResources,
	)
	ContainersCustomizer func(
		[]v1.Container,
	)
	PodCustomizer func(
		*v1.PodSpec,
	)
	ServiceCustomizer func(
		*v1.ServiceSpec,
	)
	PortsCustomizer func(
		servicePorts []v1.ServicePort,
		podPorts ...[]v1.ContainerPort,
	)
)

// Global defaults
const (
	DEFAULT_REPLICAS = int32(1)
	DEFAULT_PORT     = int32(80)
)

const (
	// Deployment is the default kind of general workload, this is what you most likely need to use
	Deployment = iota
	// ReplicaSet is a lower-level kind for a general workload, it's the same as KindDeployment, expcept it doesn't support rolloouts
	ReplicaSet
	// DaemonSet
	DaemonSet
	// StatefullSet
	StatefullSet
	Service
	ConfigMap
	Secret
)

type AppComponentResources struct {
	deployment *v1beta1.Deployment
	service    *v1.Service
	manifest   AppComponent
}
