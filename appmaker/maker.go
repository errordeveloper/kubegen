package appmaker

import (
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
// for more trivial things (like hostNetwork) we have setters and getters
type AppComponentOpts struct {
	PrometheusPath   string
	PrometheusScrape bool
	// WithoutPorts implies ExcludeService and StandardProbes
	WithoutPorts                   bool
	WithoutStandardProbes          bool
	WithoutStandardSecurityContext bool
	HealthPath                     string
	LivenessPath                   string
	// XXX we can add these here, but may be they belong elsewhere?
	//WithProbes interface{}
	//WithSecurityContext interface{}
	// WithoutService disables building of the service
	WithoutService bool
}

type AppComponent struct {
	Image    string
	Name     string
	Port     int32
	Replicas *int32
	Opts     AppComponentOpts
	Env      map[string]string
	// It's probably okay for now, but we'd eventually want to
	// inherit properties defined outside of the AppComponent struct,
	// that it anything we'd use setters and getters for, so we might
	// want to figure out intermediate struct or just write more
	// some tests to see how things would work without that...
	BasedOn *AppComponent
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

func (i *AppComponent) GetNameAndLabels() (string, map[string]string) {
	var name string

	imageParts := strings.Split(strings.Split(i.Image, ":")[0], "/")
	name = imageParts[len(imageParts)-1]

	if i.Name != "" {
		name = i.Name
	}

	labels := map[string]string{"name": name}

	return name, labels
}

func (i *AppComponent) GetMeta() kapi.ObjectMeta {
	name, labels := i.GetNameAndLabels()
	return kapi.ObjectMeta{
		Name:   name,
		Labels: labels,
	}
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

func (i *AppComponent) getPort(params AppParams) int32 {
	if i.Port != 0 {
		return i.Port
	}
	return params.DefaultPort
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

func (i *AppComponent) MakePod(params AppParams) *kapi.PodTemplateSpec {
	name, labels := i.GetNameAndLabels()
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

func (i *AppComponentResourcePair) SetHostNetwork(bool) AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) GetHostNetwork() AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) SetHostPID(bool) AppComponentResourcePair {
	return *i
}

func (i *AppComponentResourcePair) GetHostPID() AppComponentResourcePair {
	return *i
}

func (i *AppComponent) MakeDeployment(params AppParams, pod *kapi.PodTemplateSpec) *kext.Deployment {
	if pod == nil {
		return nil
	}

	meta := i.GetMeta()

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
	meta := i.GetMeta()

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
	pod := i.MakePod(params)

	return AppComponentResourcePair{
		i.MakeDeployment(params, pod),
		i.MakeService(params),
	}
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
