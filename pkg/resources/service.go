package resources

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/ulule/deepcopier"
)

func (i Service) ToObject() (runtime.Object, error) {
	obj, err := i.Convert()
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *Service) Convert() (*v1.Service, error) {
	meta := i.Metadata.Convert(i.Name)

	serviceSpec := v1.ServiceSpec{
		Ports: []v1.ServicePort{},
	}

	deepcopier.Copy(i).To(&serviceSpec)

	if len(i.Selector) == 0 {
		serviceSpec.Selector = meta.Labels
	} else {
		serviceSpec.Selector = i.Selector
	}

	// TODO validate ports are defined withing the group?
	// TODO validate labels for a given selector are defined withing the group?
	for _, port := range i.Ports {
		p := v1.ServicePort{
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
			}
		} else {
			return nil, fmt.Errorf("unable to define ports for service %q – either port or port name must be set", i.Name)
		}
		serviceSpec.Ports = append(serviceSpec.Ports, p)
	}

	service := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
		Spec:       serviceSpec,
	}

	return &service, nil
}
