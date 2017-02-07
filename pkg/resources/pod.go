package resources

import (
	"fmt"

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

func (i *Container) Convert(volumes []v1.Volume) (*v1.Container, error) {
	var err error

	container := v1.Container{Name: i.Name, Image: i.Image}

	deepcopier.Copy(i).To(&container)

	i.maybeAddEnvVars(&container)

	// you'd think the types could be simply converted,
	// but it turns out they won't because tags are different...
	// Fortunatelly, this has changed in Go1.8!
	for _, port := range i.Ports {
		container.Ports = append(container.Ports, v1.ContainerPort(port))
	}

	knownVolumes := make(map[string]bool)
	for _, volume := range volumes {
		knownVolumes[volume.Name] = true
	}

	for _, volumeMount := range i.Mounts {
		if _, ok := knownVolumes[volumeMount.Name]; !ok {
			return nil, fmt.Errorf("unable to mount volume %q in the container %q – no such volume defined for this pod", volumeMount.Name, i.Name)
		}
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount(volumeMount))
	}

	if i.LivenessProbe != nil {
		container.LivenessProbe, err = i.LivenessProbe.Convert(i.Ports)
		if err != nil {
			return nil, fmt.Errorf("unable to add livenes probe to the container %q – %v", i.Name, err)
		}
	}

	if i.ReadinessProbe != nil {
		container.ReadinessProbe, err = i.ReadinessProbe.Convert(i.Ports)
		if err != nil {
			return nil, fmt.Errorf("unable to add readines probe to the container %q – %v", i.Name, err)
		}
	}

	if i.SecurityContext != nil {
		container.SecurityContext = &v1.SecurityContext{}
		deepcopier.Copy(i.SecurityContext).To(container.SecurityContext)
	}

	resourceRequirements, err := i.Resources.Convert()
	if err != nil {
		return nil, fmt.Errorf("unable to add resource requirements for container %q – %v", i.Name, err)
	}
	container.Resources = *resourceRequirements

	return &container, nil
}

func (i *ResourceRequirements) Convert() (*v1.ResourceRequirements, error) {
	var err error

	resourceRequirements := v1.ResourceRequirements{}
	if len(i.Limits) > 0 {
		resourceRequirements.Limits = make(v1.ResourceList)
		for k, v := range i.Limits {
			resourceRequirements.Limits[v1.ResourceName(k)], err = resource.ParseQuantity(v)
			if err != nil {
				return nil, fmt.Errorf("cannot set resource limit, value for %q does not parse – %v", k, err)
			}
		}
	}
	if len(i.Requests) > 0 {
		resourceRequirements.Requests = make(v1.ResourceList)
		for k, v := range i.Requests {
			resourceRequirements.Requests[v1.ResourceName(k)], err = resource.ParseQuantity(v)
			if err != nil {
				return nil, fmt.Errorf("cannot set resource limit, value for %q does not parse – %v", k, err)
			}
		}
	}
	return &resourceRequirements, nil
}

func (i *Probe) Convert(ports []ContainerPort) (*v1.Probe, error) {
	probe := v1.Probe{Handler: v1.Handler{}}

	deepcopier.Copy(i).To(&probe)

	var defaultPort func() intstr.IntOrString

	missingPortsError := fmt.Errorf("cannot define a probe without ports")

	// pick the first port by default
	if len(ports) > 0 {
		defaultPort = func() intstr.IntOrString {
			return intstr.FromString(ports[0].Name)
		}
	}

	whichHandler := exclusiveNonNil(i.Handler.Exec, i.Handler.HTTPGet, i.Handler.TCPSocket)
	if whichHandler != nil {
		switch *whichHandler {
		case 0:
			a := v1.ExecAction(*i.Handler.Exec)
			probe.Handler.Exec = &a
		case 1:
			if len(ports) == 0 {
				return nil, missingPortsError
			}

			a := v1.HTTPGetAction{Port: defaultPort()}
			h := i.Handler.HTTPGet

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
			if len(ports) == 0 {
				return nil, missingPortsError
			}

			a := v1.TCPSocketAction{Port: defaultPort()}
			h := i.Handler.TCPSocket

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
	} else {
		return nil, fmt.Errorf("one probe handler must be defined, none or too many given")
	}

	return &probe, nil
}

func (i *Volume) Convert() (*v1.Volume, error) {
	volume := v1.Volume{Name: i.Name}

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
	} else {
		return nil, fmt.Errorf("one volume source must be defined, none or too many given")
	}

	return &volume, nil
}

func MakePod(parentMeta metav1.ObjectMeta, spec Pod) (*v1.PodTemplateSpec, error) {
	meta := metav1.ObjectMeta{
		Labels:      parentMeta.Labels,
		Annotations: spec.Annotations,
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{},
		Volumes:    []v1.Volume{},
	}

	deepcopier.Copy(spec).To(&podSpec)

	for n, volume := range spec.Volumes {
		v, err := volume.Convert()
		if err != nil {
			return nil, fmt.Errorf("error adding volume #%d/%s – %v", n, volume.Name, err)
		}
		podSpec.Volumes = append(podSpec.Volumes, *v)
	}

	for n, initContainer := range spec.InitContainers {
		c, err := initContainer.Convert(podSpec.Volumes)
		if err != nil {
			return nil, fmt.Errorf("error adding init container #%d/%s – %v", n, initContainer.Name, err)
		}
		podSpec.InitContainers = append(podSpec.InitContainers, *c)
	}

	for n, container := range spec.Containers {
		c, err := container.Convert(podSpec.Volumes)
		if err != nil {
			return nil, fmt.Errorf("error adding container #%d/%s – %v", n, container.Name, err)
		}
		podSpec.Containers = append(podSpec.Containers, *c)
	}

	if spec.SecurityContext != nil {
		s := v1.PodSecurityContext(*spec.SecurityContext)
		podSpec.SecurityContext = &s
	}

	pod := v1.PodTemplateSpec{
		ObjectMeta: meta,
		Spec:       podSpec,
	}

	return &pod, nil
}
