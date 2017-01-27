package appmaker

import (
	"sort"
	"strings"

	unversioned "k8s.io/client-go/pkg/api/unversioned" // Should eventually migrate to "k8s.io/apimachinery/pkg/apis/meta/v1"?
	kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

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

type AppComponent struct {
	Image    string
	Name     string            `json:",omitempty"`
	Port     int32             `json:",omitempty"`
	Replicas *int32            `json:",omitempty"`
	Opts     AppComponentOpts  `json:",omitempty"`
	Env      map[string]string `json:",omitempty"`
	// Deployment, DaemonSet, StatefullSet ...etc
	Kind int `json:",omitempty"`
	// It's probably okay for now, but we'd eventually want to
	// inherit properties defined outside of the AppComponent struct,
	// that it anything we'd use setters and getters for, so we might
	// want to figure out intermediate struct or just write more
	// some tests to see how things would work without that...
	BasedOn            *AppComponent        `json:",omitempty"`
	Customize          GeneralCustomizer    `json:"-"`
	CustomizeCotainers ContainersCustomizer `json:"-"`
	CustomizePod       PodCustomizer        `json:"-"`
	CustomizeService   ServiceCustomizer    `json:"-"`
	CustomizePorts     PortsCustomizer      `json:"-"`
}

type (
	GeneralCustomizer func(
		*AppComponentResources,
	)
	ContainersCustomizer func(
		[]kapi.Container,
	)
	PodCustomizer func(
		*kapi.PodSpec,
	)
	ServiceCustomizer func(
		*kapi.ServiceSpec,
	)
	PortsCustomizer func(
		servicePorts []kapi.ServicePort,
		podPorts ...[]kapi.ContainerPort,
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

// Everything we want to controll per-app
type AppParams struct {
	Namespace              string
	DefaultReplicas        int32
	DefaultPort            int32
	StandardLivenessProbe  *kapi.Probe
	StandardReadinessProbe *kapi.Probe
}

type App struct {
	GroupName  string
	Components []AppComponent
}

type AppComponentResources struct {
	deployment *kext.Deployment
	service    *kapi.Service
	manifest   AppComponent
}

type AppMaker interface {
	MarshalToJSON() ([]map[int][]byte, error)
	//MakeAll(params AppParams) *AppComponentResources
}

func (i *AppComponent) getNameAndLabels() (string, map[string]string) {
	var name string

	imageParts := strings.Split(strings.Split(i.Image, ":")[0], "/")
	name = imageParts[len(imageParts)-1]

	if i.Name != "" {
		name = i.Name
	}

	labels := map[string]string{"name": name}

	return name, labels
}

func (i *AppComponent) getMeta() kapi.ObjectMeta {
	name, labels := i.getNameAndLabels()
	return kapi.ObjectMeta{
		Name:   name,
		Labels: labels,
	}
}

func (i *AppComponent) getPort(params AppParams) int32 {
	if i.Port != 0 {
		return i.Port
	}
	return params.DefaultPort
}

func (i *AppComponent) maybeAddEnvVars(params AppParams, container *kapi.Container) {
	keys := []string{}
	for k, _ := range i.Env {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return
	}

	sort.Strings(keys)

	env := []kapi.EnvVar{}
	for _, j := range keys {
		for k, v := range i.Env {
			if k == j {
				env = append(env, kapi.EnvVar{Name: k, Value: v})
			}
		}
	}
	container.Env = env
}

func (i *AppComponent) maybeAddProbes(params AppParams, container *kapi.Container) {
	if i.Opts.WithoutStandardProbes {
		return
	}
	port := intstr.FromInt(int(i.getPort(params)))

	container.ReadinessProbe = &kapi.Probe{
		PeriodSeconds:       3,
		InitialDelaySeconds: 180,
		Handler: kapi.Handler{
			HTTPGet: &kapi.HTTPGetAction{
				Path: "/health",
				Port: port,
			},
		},
	}
	container.LivenessProbe = &kapi.Probe{
		PeriodSeconds:       3,
		InitialDelaySeconds: 300,
		Handler: kapi.Handler{
			HTTPGet: &kapi.HTTPGetAction{
				Path: "/health",
				Port: port,
			},
		},
	}
}

func (i *AppComponent) MakeContainer(params AppParams, name string) kapi.Container {
	container := kapi.Container{Name: name, Image: i.Image}

	i.maybeAddEnvVars(params, &container)

	if !i.Opts.WithoutPorts {
		container.Ports = []kapi.ContainerPort{{
			Name:          name,
			ContainerPort: i.getPort(params),
		}}
		i.maybeAddProbes(params, &container)
	}

	return container
}

func (i *AppComponent) MakePod(params AppParams) *kapi.PodTemplateSpec {
	name, labels := i.getNameAndLabels()
	container := i.MakeContainer(params, name)

	pod := kapi.PodTemplateSpec{
		ObjectMeta: kapi.ObjectMeta{
			Labels: labels,
		},
		Spec: kapi.PodSpec{
			Containers: []kapi.Container{container},
		},
	}

	if i.CustomizePod != nil {
		i.CustomizePod(&pod.Spec)
	}

	if i.CustomizeCotainers != nil {
		i.CustomizeCotainers(pod.Spec.Containers)
	}

	return &pod
}

func (i *AppComponent) MakeDeployment(params AppParams, pod *kapi.PodTemplateSpec) *kext.Deployment {
	if pod == nil {
		return nil
	}

	meta := i.getMeta()

	replicas := params.DefaultReplicas

	if i.Replicas != nil {
		replicas = *i.Replicas
	}

	deploymentSpec := kext.DeploymentSpec{
		Replicas: &replicas,
		Selector: &unversioned.LabelSelector{MatchLabels: meta.Labels},
		Template: *pod,
	}

	deployment := &kext.Deployment{
		ObjectMeta: meta,
		Spec:       deploymentSpec,
	}

	if params.Namespace != "" {
		deployment.ObjectMeta.Namespace = params.Namespace
	}

	return deployment
}

func (i *AppComponent) MakeService(params AppParams) *kapi.Service {
	meta := i.getMeta()

	port := kapi.ServicePort{Port: i.getPort(params)}
	if i.Port != 0 {
		port.Port = i.Port
	}

	service := &kapi.Service{
		ObjectMeta: meta,
		Spec: kapi.ServiceSpec{
			Ports:    []kapi.ServicePort{port},
			Selector: meta.Labels,
		},
	}

	if i.CustomizeService != nil {
		i.CustomizeService(&service.Spec)
	}

	return service
}

func (i *AppComponent) MakeAll(params AppParams) *AppComponentResources {
	resources := AppComponentResources{manifest: *i}

	switch i.Kind {
	case Deployment:
		resources.deployment = i.MakeDeployment(params, i.MakePod(params))
	}

	if !i.Opts.WithoutService {
		resources.service = i.MakeService(params)
	}

	if i.CustomizePorts != nil && !i.Opts.WithoutPorts {
		containers := resources.Deployment().Spec.Template.Spec.Containers
		containerPorts := make([][]kapi.ContainerPort, len(containers))
		for n, container := range containers {
			containerPorts[n] = container.Ports
		}
		i.CustomizePorts(
			nil, //resources.Service().Spec.Ports,
			containerPorts...,
		)
	}

	if i.Customize != nil {
		i.Customize(&resources)
	}

	return &resources
}

func (i *AppComponentResources) AppendContainer(container kapi.Container) AppComponentResources {
	containers := &i.Deployment().Spec.Template.Spec.Containers
	*containers = append(*containers, container)
	return *i
}

func (i *AppComponentResources) MountDataVolume() AppComponentResources {
	// TODO append to volumes and volume mounts based on few simple parameters
	// when user uses more then one container, they will have to do it in a low-level way
	// secrets and config maps would be handled separatelly, so we call this MountDataVolume()
	// and not something else
	return *i
}

func (i *AppComponentResources) WithSecret(secretData interface{}) AppComponentResources {
	return *i
}

func (i *AppComponentResources) WithConfig(configMapData interface{}) AppComponentResources {
	return *i
}

func (i *AppComponentResources) WithExtraLabels(map[string]string) AppComponentResources {
	return *i
}

func (i *AppComponentResources) WithExtraAnnotations(map[string]string) AppComponentResources {
	return *i
}

func (i *AppComponentResources) WithExtraPorts(interface{}) AppComponentResources {
	// TODO May be this should be a customizer, i.e. PortsCustomizer closure
	return *i
}

func (i *AppComponentResources) UseHostNetwork() AppComponentResources {
	return *i
}

func (i *AppComponentResources) UseHostPID() AppComponentResources {
	return *i
}

func (i *AppComponentResources) Deployment() *kext.Deployment {
	return i.deployment
}

func (i *AppComponentResources) Service() *kapi.Service {
	return i.service
}

// TODO: params argument
func (i *App) MakeAll() []*AppComponentResources {
	params := AppParams{
		Namespace:       i.GroupName,
		DefaultReplicas: DEFAULT_REPLICAS,
		DefaultPort:     DEFAULT_PORT,
		// standardSecurityContext
		// standardTmpVolume?
	}

	list := []*AppComponentResources{}

	for _, service := range i.Components {
		list = append(list, service.MakeAll(params))
	}

	return list
}
