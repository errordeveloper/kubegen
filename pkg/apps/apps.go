package apps

import (
	_ "fmt"

	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"

	"github.com/imdario/mergo"

	"github.com/errordeveloper/kubegen/pkg/resources"
)

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

func (i *AppComponent) getPort(params AppParams) int32 {
	if i.Port != 0 {
		return i.Port
	}
	return params.DefaultPort
}

func (i *AppComponent) maybeAddEnvVars(container *resources.Container) {
	if len(i.Env) == 0 {
		return
	}

	if container.Env == nil {
		container.Env = make(map[string]string)
	}

	for k, v := range i.Env {
		container.Env[k] = v
	}
}

func (i *AppComponent) maybeUseCommonEnvVars(params AppParams) {
	if len(i.CommonEnv) == 0 {
		return
	}

	if i.Env == nil {
		i.Env = make(map[string]string)
	}

	for _, j := range i.CommonEnv {
		if v, ok := params.commonEnv[j]; ok {
			i.Env[j] = v
		}
	}
}

func (i *AppComponent) maybeAddProbes(params AppParams, container *v1.Container) {
	if i.Opts.WithoutStandardProbes {
		return
	}
	port := intstr.FromInt(int(i.getPort(params)))

	container.ReadinessProbe = &v1.Probe{
		PeriodSeconds:       3,
		InitialDelaySeconds: 180,
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/health",
				Port: port,
			},
		},
	}
	container.LivenessProbe = &v1.Probe{
		PeriodSeconds:       3,
		InitialDelaySeconds: 300,
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/health",
				Port: port,
			},
		},
	}
}

func (i *AppComponent) MakeContainer(params AppParams, name string) *resources.Container {
	container := resources.Container{Name: name, Image: i.Image}

	i.maybeUseCommonEnvVars(params)
	i.maybeAddEnvVars(&container)

	if !i.Opts.WithoutPorts {
		container.Ports = []resources.ContainerPort{{
			Name:          name,
			ContainerPort: i.getPort(params),
		}}
		//i.maybeAddProbes(params, &container)
	}

	return &container
}

func (i *AppComponent) MakeDeployment(params AppParams) *v1beta1.Deployment {
	name, labels := i.getNameAndLabels()

	deployment := resources.Deployment{
		Name:       name,
		Metadata:   resources.Metadata{Labels: labels},
		Replicas:   params.DefaultReplicas,
		Containers: []resources.Container{*i.MakeContainer(params, name)},
	}

	if i.Replicas != nil {
		deployment.Replicas = *i.Replicas
	}

	deploymentObj := deployment.Convert()

	if params.Namespace != "" {
		deploymentObj.ObjectMeta.Namespace = params.Namespace
	}

	return deploymentObj
}

func (i *AppComponent) MakeService(params AppParams) *v1.Service {
	name, labels := i.getNameAndLabels()

	port := resources.ServicePort{
		Name: name,
		Port: i.getPort(params),
	}

	if i.Port != 0 {
		port.Port = i.Port
	}

	service := resources.Service{
		Name:     name,
		Metadata: resources.Metadata{Labels: labels},
		Selector: labels,
		Ports:    []resources.ServicePort{port},
	}

	serviceObj := service.Convert()

	if params.Namespace != "" {
		serviceObj.ObjectMeta.Namespace = params.Namespace
	}

	return serviceObj
}

func (i *AppComponent) MakeAll(params AppParams) *AppComponentResources {
	resources := AppComponentResources{}

	if i.BasedOnNamedTemplate != "" {
		if template, ok := params.templates[i.BasedOnNamedTemplate]; ok {
			i.basedOn = &template
		}
	}

	if i.basedOn != nil {
		if i.Env == nil {
			i.Env = make(map[string]string)
		}
		base := *i.basedOn
		if err := mergo.Merge(&base, *i); err != nil {
			panic(err)
		}
		if err := mergo.Merge(i, base); err != nil {
			panic(err)
		}
	}

	resources.manifest = *i

	switch i.Kind {
	case Deployment:
		resources.deployment = i.MakeDeployment(params)
	}

	pod := resources.getPod()

	if !i.Opts.WithoutService {
		resources.service = i.MakeService(params)
	}

	if i.Flavor != "" {
		if fn, ok := Flavors[i.Flavor]; ok {
			fn(&resources)
		}
	}

	if i.CustomizePorts != nil && !i.Opts.WithoutPorts {
		ports := make([][]v1.ContainerPort, len(pod.Containers))
		for n, container := range pod.Containers {
			ports[n] = container.Ports
		}
		i.CustomizePorts(
			resources.service.Spec.Ports,
			ports...,
		)
	}

	if i.CustomizeCotainers != nil {
		i.CustomizeCotainers(pod.Containers)
	}

	if i.CustomizePod != nil {
		i.CustomizePod(pod)
	}

	if i.CustomizeService != nil {
		i.CustomizeService(&resources.service.Spec)
	}

	if i.Customize != nil {
		i.Customize(&resources)
	}

	return &resources
}

func (i *AppComponent) MakeList(params AppParams) *api.List {
	resources := i.MakeAll(params)

	list := &api.List{}
	switch i.Kind {
	case Deployment:
		list.Items = append(list.Items, runtime.Object(resources.deployment))
	}

	if resources.service != nil {
		list.Items = append(list.Items, runtime.Object(resources.service))

	}

	return list
}

func (i *AppComponentResources) AppendContainer(container v1.Container) AppComponentResources {
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
	// TODO May be this should be a customizer, i.e. it'd basically create a PortsCustomizer closure and return it
	return *i
}

func (i *AppComponentResources) UseHostNetwork() AppComponentResources {
	return *i
}

func (i *AppComponentResources) UseHostPID() AppComponentResources {
	return *i
}

func (i *AppComponentResources) Deployment() *v1beta1.Deployment {
	return i.deployment
}

func (i *AppComponentResources) Service() *v1.Service {
	return i.service
}

func (i *AppComponentResources) getPod() *v1.PodSpec {
	switch i.manifest.Kind {
	case Deployment:
		return &i.deployment.Spec.Template.Spec
	default:
		return nil
	}

}

func (i *AppComponentResources) getContainers() []v1.Container {
	switch i.manifest.Kind {
	case Deployment:
		return i.deployment.Spec.Template.Spec.Containers
	default:
		return nil
	}

}

func (i *App) makeDefaultParams() AppParams {
	params := AppParams{
		Namespace:       i.GroupName,
		DefaultReplicas: DEFAULT_REPLICAS,
		DefaultPort:     DEFAULT_PORT,
		templates:       make(map[string]AppComponent),
		// standardSecurityContext
		// standardTmpVolume?
	}

	for _, template := range i.Templates {
		t := &AppComponent{
			Image: template.Image,
			Env:   make(map[string]string),
		}
		if err := mergo.Merge(t, template.AppComponent); err != nil {
			panic(err)
		}
		params.templates[template.TemplateName] = *t
	}

	if len(i.CommonEnv) != 0 {
		params.commonEnv = i.CommonEnv
	}

	return params
}

// TODO: params argument
func (i *App) MakeAll() []*AppComponentResources {
	params := i.makeDefaultParams()

	list := []*AppComponentResources{}
	for _, component := range i.Components {
		list = append(list, component.MakeAll(params))
	}

	for _, component := range i.ComponentsFromImages {
		c := &AppComponent{
			Image: component.Image,
			Env:   make(map[string]string),
		}
		if err := mergo.Merge(c, component.AppComponent); err != nil {
			panic(err)
		}
		list = append(list, c.MakeAll(params))
	}

	for _, component := range i.ComponentsFromTemplates {
		// TODO we may want to return an error if template referenced here is not defined
		c := &AppComponent{
			Image:                component.Image,
			BasedOnNamedTemplate: component.TemplateName,
			Env:                  make(map[string]string),
		}
		if err := mergo.Merge(c, component.AppComponent); err != nil {
			panic(err)
		}
		list = append(list, c.MakeAll(params))
	}

	return list
}

func (i *App) MakeList() *api.List {
	params := i.makeDefaultParams()

	list := &api.List{}
	for _, component := range i.Components {
		list.Items = append(list.Items, component.MakeList(params).Items...)
	}

	for _, component := range i.ComponentsFromImages {
		c := &AppComponent{
			Image: component.Image,
			Env:   make(map[string]string),
		}
		if err := mergo.Merge(c, component.AppComponent); err != nil {
			panic(err)
		}
		list.Items = append(list.Items, c.MakeList(params).Items...)
	}

	for _, component := range i.ComponentsFromTemplates {
		// TODO we may want to return an error if template referenced here is not defined
		c := &AppComponent{
			Image:                component.Image,
			BasedOnNamedTemplate: component.TemplateName,
			Env:                  make(map[string]string),
		}
		if err := mergo.Merge(c, component.AppComponent); err != nil {
			panic(err)
		}
		list.Items = append(list.Items, c.MakeList(params).Items...)
	}

	return list
}
