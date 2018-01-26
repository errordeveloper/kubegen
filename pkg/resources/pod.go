package resources

import (
	"fmt"

	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/ulule/deepcopier"
)

func (i *Container) maybeAddEnvVars(container *corev1.Container) {
	if len(i.Env) == 0 {
		return
	}

	keys := []string{}
	for k, _ := range i.Env {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	env := []corev1.EnvVar{}
	for _, j := range keys {
		for k, v := range i.Env {
			if k == j {
				env = append(env, corev1.EnvVar{Name: k, Value: v})
			}
		}
	}
	container.Env = env
}

func (i *Container) Convert(volumes []corev1.Volume) (*corev1.Container, error) {
	var err error

	container := corev1.Container{Name: i.Name, Image: i.Image}

	deepcopier.Copy(i).To(&container)

	i.maybeAddEnvVars(&container)

	// Ensure that any singleton unnamed port inherits the
	// name from the container's name
	if len(i.Ports) == 1 {
		if i.Ports[0].Name == "" {
			i.Ports[0].Name = container.Name
		}
	}

	// you'd think the types could be simply converted,
	// but it turns out they won't because tags are different...
	// Fortunatelly, this has changed in Go1.8!
	for _, port := range i.Ports {
		container.Ports = append(container.Ports, corev1.ContainerPort(port))
	}

	knownVolumes := make(map[string]bool)
	for _, volume := range volumes {
		knownVolumes[volume.Name] = true
	}

	for _, volumeMount := range i.VolumeMounts {
		if _, ok := knownVolumes[volumeMount.Name]; !ok {
			return nil, fmt.Errorf("unable to mount volume %q in the container %q – no such volume defined for this pod", volumeMount.Name, i.Name)
		}
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount(volumeMount))
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
		container.SecurityContext = &corev1.SecurityContext{}
		deepcopier.Copy(i.SecurityContext).To(container.SecurityContext)
	}

	resourceRequirements, err := i.Resources.Convert()
	if err != nil {
		return nil, fmt.Errorf("unable to add resource requirements for container %q – %v", i.Name, err)
	}
	container.Resources = *resourceRequirements

	return &container, nil
}

func (i *ResourceRequirements) Convert() (*corev1.ResourceRequirements, error) {
	var err error

	resourceRequirements := corev1.ResourceRequirements{}
	if len(i.Limits) > 0 {
		resourceRequirements.Limits = make(corev1.ResourceList)
		for k, v := range i.Limits {
			resourceRequirements.Limits[corev1.ResourceName(k)], err = resource.ParseQuantity(v)
			if err != nil {
				return nil, fmt.Errorf("cannot set resource limit, value for %q does not parse – %v", k, err)
			}
		}
	}
	if len(i.Requests) > 0 {
		resourceRequirements.Requests = make(corev1.ResourceList)
		for k, v := range i.Requests {
			resourceRequirements.Requests[corev1.ResourceName(k)], err = resource.ParseQuantity(v)
			if err != nil {
				return nil, fmt.Errorf("cannot set resource limit, value for %q does not parse – %v", k, err)
			}
		}
	}
	return &resourceRequirements, nil
}

func (i *Probe) Convert(ports []ContainerPort) (*corev1.Probe, error) {
	probe := corev1.Probe{Handler: corev1.Handler{}}

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
			a := corev1.ExecAction(*i.Handler.Exec)
			probe.Handler.Exec = &a
		case 1:
			if len(ports) == 0 {
				return nil, missingPortsError
			}

			a := corev1.HTTPGetAction{Port: defaultPort()}
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
					a.HTTPHeaders = append(a.HTTPHeaders, corev1.HTTPHeader{Name: k, Value: v})
				}
			}

			deepcopier.Copy(h).To(&a)

			probe.Handler.HTTPGet = &a
		case 2:
			if len(ports) == 0 {
				return nil, missingPortsError
			}

			a := corev1.TCPSocketAction{Port: defaultPort()}
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

func (i *Volume) Convert() (*corev1.Volume, error) {
	volume := corev1.Volume{Name: i.Name}

	whichVolumeSource := exclusiveNonNil(i.HostPath, i.EmptyDir, i.Secret, i.ConfigMap, i.PersistentVolumeClaim)
	if whichVolumeSource != nil {
		switch *whichVolumeSource {
		case 0:
			s := corev1.HostPathVolumeSource(*i.VolumeSource.HostPath)
			volume.VolumeSource.HostPath = &s
		case 1:
			s := corev1.EmptyDirVolumeSource(*i.VolumeSource.EmptyDir)
			volume.VolumeSource.EmptyDir = &s
		case 2:
			s := corev1.SecretVolumeSource{}
			deepcopier.Copy(i.VolumeSource.Secret).To(&s)
			if s.SecretName == "" {
				s.SecretName = i.Name
			}
			volume.VolumeSource.Secret = &s
		case 3:
			s := corev1.ConfigMapVolumeSource{}
			deepcopier.Copy(i.VolumeSource.ConfigMap).To(&s)
			if s.Name == "" {
				s.Name = i.Name
			}
			volume.VolumeSource.ConfigMap = &s
		case 4:
			volume.VolumeSource.PersistentVolumeClaim = i.PersistentVolumeClaim
		}
	} else {
		return nil, fmt.Errorf("one volume source must be defined, none or too many given")
	}

	return &volume, nil
}

func MakePod(parentMeta metav1.ObjectMeta, spec Pod) (*corev1.PodTemplateSpec, error) {
	meta := metav1.ObjectMeta{
		Labels:      parentMeta.Labels,
		Annotations: spec.Annotations,
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{},
		Volumes:    []corev1.Volume{},
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
		s := corev1.PodSecurityContext(*spec.SecurityContext)
		podSpec.SecurityContext = &s
	}

	pod := corev1.PodTemplateSpec{
		ObjectMeta: meta,
		Spec:       podSpec,
	}

	return &pod, nil
}
