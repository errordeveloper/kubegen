package appmaker

import (
	"encoding/json"
	_ "fmt"
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
	// Deployment, DaemonSet, StatefullSet ...etc
	Kind             int    `json:",omitempty"`
	PrometheusPath   string `json:",omitempty"`
	PrometheusScrape bool   `json:",omitempty"`

	// WithoutPorts implies ExcludeService and StandardProbes
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
	// It's probably okay for now, but we'd eventually want to
	// inherit properties defined outside of the AppComponent struct,
	// that it anything we'd use setters and getters for, so we might
	// want to figure out intermediate struct or just write more
	// some tests to see how things would work without that...
	BasedOn   *AppComponent                   `json:",omitempty"`
	Customize func(*AppComponentResourcePair) `json:"-"`
}

// Global defaults
const (
	DEFAULT_REPLICAS = int32(1)
	DEFAULT_PORT     = int32(80)
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
	Name  string
	Group []AppComponent
}

// TODO figure out how to use kapi.List here, if we can
// TODO find a way to use something other then Deployment
// e.g. StatefullSet or DaemonSet, also attach a ConfigMap
// or a secret or several of those things
type AppComponentResourcePair struct {
	Deployment *kext.Deployment
	Service    *kapi.Service
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

	return &pod
}

func (i *AppComponentResourcePair) AppendContainer(container kapi.Container) AppComponentResourcePair {
	containers := &i.Deployment.Spec.Template.Spec.Containers
	*containers = append(*containers, container)
	return *i
}

func (i *AppComponentResourcePair) MountDataVolume() AppComponentResourcePair {
	// TODO append to volumes and volume mounts based on few simple parameters
	// when user uses more then one container, they will have to do it in a low-level way
	// secrets and config maps would be handled separatelly, so we call this MountDataVolume()
	// and not something else
	return *i
}

func (i *AppComponentResourcePair) WithSecret(secretData interface{}) AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) WithConfig(configMapData interface{}) AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) WithExtraLabels(map[string]string) AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) WithExtraAnnotations(map[string]string) AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) UseHostNetwork() AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) UseHostPID(bool) AppComponentResourcePair {
	return *i
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

	return service
}

func (i *AppComponent) MakeAll(params AppParams) AppComponentResourcePair {
	resources := AppComponentResourcePair{
		i.MakeDeployment(params, i.MakePod(params)),
		i.MakeService(params),
	}

	if i.Customize != nil {
		i.Customize(&resources)
	}

	return resources
}

func (i *App) MakeAll() []AppComponentResourcePair {
	params := AppParams{
		Namespace:       i.Name,
		DefaultReplicas: DEFAULT_REPLICAS,
		DefaultPort:     DEFAULT_PORT,
		// standardSecurityContext
		// standardTmpVolume?
	}

	list := []AppComponentResourcePair{}

	for _, service := range i.Group {
		list = append(list, service.MakeAll(params))
	}

	return list
}

func NewFromJSON(manifest []byte) (*App, error) {
	app := &App{}
	if err := json.Unmarshal(manifest, app); err != nil {
		return nil, err
	}
	return app, nil
}
