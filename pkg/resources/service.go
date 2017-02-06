package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/ulule/deepcopier"
)

func (i *Service) Convert() *v1.Service {
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
		} // TODO: should error if both are set
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

	return &service
}
