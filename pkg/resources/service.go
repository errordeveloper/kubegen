package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/ulule/deepcopier"
)

func (i Service) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *Group) findPortByName(serviceName, portName string) (*ContainerPort, error) {
	allPorts := []ContainerPort{}
	// TODO we should match a selector, but it is not set before conversion takes place
	for _, component := range i.Deployments {
		if component.Name == serviceName {
			for _, container := range component.Pod.Containers {
				allPorts = append(allPorts, container.Ports...)
			}
		}
	}
	for _, component := range i.ReplicaSets {
		if component.Name == serviceName {
			for _, container := range component.Pod.Containers {
				allPorts = append(allPorts, container.Ports...)
			}
		}

	}
	for _, component := range i.DaemonSets {
		if component.Name == serviceName {
			for _, container := range component.Pod.Containers {
				allPorts = append(allPorts, container.Ports...)
			}
		}
	}
	for _, component := range i.StatefulSets {
		if component.Name == serviceName {
			for _, container := range component.Pod.Containers {
				allPorts = append(allPorts, container.Ports...)
			}
		}
	}

	matchingPorts := []ContainerPort{}
	for _, port := range allPorts {
		if port.Name == portName {
			matchingPorts = append(matchingPorts, port)
		}
	}

	if len(matchingPorts) == 0 {
		return nil, fmt.Errorf("no port matching name %q found for Service name %q in pod controllers with the same name", portName, serviceName)
	}

	if len(matchingPorts) > 1 {
		return nil, fmt.Errorf("too many (%d) port matching name %q found for Service name %q in pod controllers with the same name", len(matchingPorts), portName, serviceName)
	}

	return &matchingPorts[0], nil
}

func (i *Service) Convert(localGroup *Group) (*corev1.Service, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	serviceSpec := corev1.ServiceSpec{
		Ports: []corev1.ServicePort{},
	}

	deepcopier.Copy(i).To(&serviceSpec)

	if len(i.Selector) == 0 {
		serviceSpec.Selector = meta.Labels
	} else {
		serviceSpec.Selector = i.Selector
	}

	nodePortIsSet := false

	// TODO validate ports are defined withing the group?
	// TODO validate labels for a given selector are defined withing the group?
	for _, port := range i.Ports {
		p := corev1.ServicePort{
			Name:     port.Name,
			Port:     port.Port,
			NodePort: port.NodePort,
			// default to taget port with the same name
			TargetPort: intstr.FromString(port.Name),
		}
		if !(port.TargetPort != 0 && port.TargetPortName != "") {
			// overide the default target port
			if port.TargetPort != 0 {
				p.TargetPort = intstr.FromInt(int(port.TargetPort))
				if port.Port == 0 {
					p.Port = port.TargetPort
				}
			}
			if port.TargetPortName != "" {
				p.TargetPort = intstr.FromString(port.TargetPortName)
				if p.Port == 0 && p.Name != "" {
					matchingPort, err := localGroup.findPortByName(i.Name, port.TargetPortName)
					if err != nil {
						return nil, err
					}
					p.Port = matchingPort.ContainerPort
				}
			}
		} else {
			return nil, fmt.Errorf("unable to define ports for service %q – either port or port name must be set", i.Name)
		}

		if p.Port == 0 && p.Name != "" {
			matchingPort, err := localGroup.findPortByName(i.Name, p.Name)
			if err != nil {
				return nil, err
			}
			p.Port = matchingPort.ContainerPort
		}

		if p.NodePort != 0 {
			nodePortIsSet = true
		}

		serviceSpec.Ports = append(serviceSpec.Ports, p)
	}

	if serviceSpec.Type == "" && nodePortIsSet {
		serviceSpec.Type = "NodePort"
	}

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
		Spec:       serviceSpec,
	}

	return &service, nil
}
