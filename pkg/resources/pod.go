package resources

import (
	"sort"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/ulule/deepcopier"
)

func (i *Container) maybeAddEnvVars(container *v1.Container) {
	if len(i.Env) == 0 {
		return
	}

	keys := []string{}
	for k, _ := range i.Env {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	env := []v1.EnvVar{}
	for _, j := range keys {
		for k, v := range i.Env {
			if k == j {
				env = append(env, v1.EnvVar{Name: k, Value: v})
			}
		}
	}
	container.Env = env
}

func (i *Container) Convert() v1.Container {
	container := v1.Container{Name: i.Name, Image: i.Image}

	deepcopier.Copy(i).To(&container)

	i.maybeAddEnvVars(&container)

	// you'd think the types could be simply converted,
	// but it turns out they won't because tags are different...
	// Fortunatelly, this has changed in Go1.8!
	//container.Ports = []v1.ContainerPort(i.Ports)
	for _, port := range i.Ports {
		container.Ports = append(container.Ports, v1.ContainerPort(port))
	}

	for _, volumeMount := range i.Mounts {
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount(volumeMount))
	}

	if i.LivenessProbe != nil {
		container.LivenessProbe = i.LivenessProbe.Convert(i.Ports)
	}

	if i.ReadinessProbe != nil {
		container.ReadinessProbe = i.ReadinessProbe.Convert(i.Ports)
	}

	if i.SecurityContext != nil {
		container.SecurityContext = &v1.SecurityContext{}
		deepcopier.Copy(i.SecurityContext).To(container.SecurityContext)
	}

	container.Resources = i.Resources.Convert()

	return container
}

func (i *ResourceRequirements) Convert() v1.ResourceRequirements {
	resourceRequirements := v1.ResourceRequirements{}
	if len(i.Limits) > 0 {
		resourceRequirements.Limits = make(v1.ResourceList)
		for k, v := range i.Limits {
			resourceRequirements.Limits[v1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if len(i.Requests) > 0 {
		resourceRequirements.Requests = make(v1.ResourceList)
		for k, v := range i.Requests {
			resourceRequirements.Requests[v1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	return resourceRequirements
}

func (i *Probe) Convert(ports []ContainerPort) *v1.Probe {
	probe := v1.Probe{Handler: v1.Handler{}}

	deepcopier.Copy(i).To(&probe)

	defaultPort := intstr.IntOrString{}

	// pick the first port by default
	if len(ports) > 0 {
		defaultPort = intstr.FromString(ports[0].Name)
	}

	whichHandler := exclusiveNonNil(i.Handler.Exec, i.Handler.HTTPGet, i.Handler.TCPSocket)
	if whichHandler != nil {
		switch *whichHandler {
		case 0:
			a := v1.ExecAction(*i.Handler.Exec)
			probe.Handler.Exec = &a
		case 1:
			a := v1.HTTPGetAction{Port: defaultPort}
			h := i.Handler.HTTPGet

			// TODO: should error if `len(ports) == 0` and none of these are set
			if !(h.Port != 0 && h.PortName != "") {
				if h.Port != 0 {
					a.Port = intstr.FromInt(int(h.Port))
				}
				if h.PortName != "" {
					a.Port = intstr.FromString(h.PortName)
				}
			}

			if len(h.HTTPHeaders) > 0 {
				for k, v := range h.HTTPHeaders {
					a.HTTPHeaders = append(a.HTTPHeaders, v1.HTTPHeader{Name: k, Value: v})
				}
			}

			deepcopier.Copy(h).To(&a)

			probe.Handler.HTTPGet = &a
		case 2:
			a := v1.TCPSocketAction{Port: defaultPort}
			h := i.Handler.TCPSocket

			// TODO: should error if `len(ports) == 0` and none of these are set
			if !(h.Port != 0 && h.PortName != "") {
				if h.Port != 0 {
					a.Port = intstr.FromInt(int(h.Port))
				}
				if h.PortName != "" {
					a.Port = intstr.FromString(h.PortName)
				}
			}

			probe.Handler.TCPSocket = &a
		}
	} // TODO error here

	return &probe
}

func (i *Volume) Convert() v1.Volume {
	volume := v1.Volume{Name: i.Name}

	// TODO error if more then one thing is set
	whichVolumeSource := exclusiveNonNil(i.HostPath, i.EmptyDir, i.Secret)
	if whichVolumeSource != nil {
		switch *whichVolumeSource {
		case 0:
			s := v1.HostPathVolumeSource(*i.VolumeSource.HostPath)
			volume.VolumeSource.HostPath = &s
		case 1:
			s := v1.EmptyDirVolumeSource(*i.VolumeSource.EmptyDir)
			volume.VolumeSource.EmptyDir = &s
		case 2:
			s := v1.SecretVolumeSource{}
			deepcopier.Copy(i.VolumeSource.Secret).To(&s)
			volume.VolumeSource.Secret = &s
		}
	}

	return volume
}

func MakePod(parentMeta metav1.ObjectMeta, spec Pod) *v1.PodTemplateSpec {
	meta := metav1.ObjectMeta{
		Labels:      parentMeta.Labels,
		Annotations: spec.Annotations,
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{},
		Volumes:    []v1.Volume{},
	}

	deepcopier.Copy(spec).To(&podSpec)

	for _, volume := range spec.Volumes {
		podSpec.Volumes = append(podSpec.Volumes, volume.Convert())
	}

	for _, initContainer := range spec.InitContainers {
		podSpec.InitContainers = append(podSpec.InitContainers, initContainer.Convert())
	}

	for _, container := range spec.Containers {
		podSpec.Containers = append(podSpec.Containers, container.Convert())
	}

	if spec.SecurityContext != nil {
		s := v1.PodSecurityContext(*spec.SecurityContext)
		podSpec.SecurityContext = &s
	}

	pod := v1.PodTemplateSpec{
		ObjectMeta: meta,
		Spec:       podSpec,
	}

	return &pod
}
