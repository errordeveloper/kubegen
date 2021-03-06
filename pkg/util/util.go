package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/printers"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/hcl"
)

func marshalToJSON(object runtime.Object) ([]byte, error) {
	// TODO consider borrowing sorting code from pkg/kubectl/sorting_printer.go
	jsprinter := printers.JSONPrinter{}
	buf := &bytes.Buffer{}
	if err := jsprinter.PrintObj(object, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func Encode(object runtime.Object, contentType string, pretty bool) ([]byte, error) {
	data, err := marshalToJSON(object)
	if err != nil {
		return nil, err
	}
	return cleanup(contentType, data, pretty)
}

func EncodeList(list *metav1.List, contentType string, pretty bool) ([]byte, error) {
	data, err := marshalToJSON(list)
	if err != nil {
		return nil, err
	}
	return cleanup(contentType, data, pretty)
}

func Decode(data []byte) (runtime.Object, error) {
	obj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), data)
	if err != nil {
		return nil, fmt.Errorf("kubegen/util: error decoding object– %v", err)
	}

	return obj, nil
}

func DumpListToFiles(list *metav1.List, contentType string) ([]string, error) {
	filenames := []string{}
	for _, item := range list.Items {
		var (
			name, filename, filenamefmt string
		)

		i := item.Object

		switch i.GetObjectKind().GroupVersionKind().Kind {
		case "Service":
			filenamefmt = "%s-svc.%s"
			name = i.(*corev1.Service).ObjectMeta.Name
		case "Deployment":
			filenamefmt = "%s-dpl.%s"
			name = i.(*appsv1.Deployment).ObjectMeta.Name
		case "ReplicaSet":
			filenamefmt = "%s-rs.%s"
			name = i.(*appsv1.ReplicaSet).ObjectMeta.Name
		case "DaemonSet":
			filenamefmt = "%s-ds.%s"
			name = i.(*appsv1.DaemonSet).ObjectMeta.Name
		case "StatefulSet":
			filenamefmt = "%s-ss.%s"
			name = i.(*appsv1.StatefulSet).ObjectMeta.Name
		}

		data, err := Encode(i, contentType, true)
		if err != nil {
			return nil, err
		}

		switch contentType {
		case "application/yaml":
			filename = fmt.Sprintf(filenamefmt, name, "yaml")
			data = append([]byte(fmt.Sprintf("# generated by kubegen\n# => %s\n---\n", filename)), data...)
		case "application/json":
			filename = fmt.Sprintf(filenamefmt, name, "yaml")
		}

		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			return nil, fmt.Errorf("kubegen/util: error writing to file %q – %v", filename, err)
		}
		filenames = append(filenames, filename)
	}

	return filenames, nil
}

func NewFromHCL(obj interface{}, data []byte) error {
	manifest, err := hcl.Parse(string(data))
	if err != nil {
		return fmt.Errorf("kubegen/util: error parsing HCL – %v", err)
	}

	if err := hcl.DecodeObject(obj, manifest); err != nil {
		return fmt.Errorf("kubegen/util: error constructing an object from HCL – %v", err)
	}

	return nil
}

func LoadObj(obj interface{}, data []byte, sourcePath string, instanceName string) error {
	var errorFmt string

	if instanceName != "" {
		errorFmt = fmt.Sprintf("kubegen/util: error loading module %q source", instanceName)
	} else {
		errorFmt = "kubegen/util: error loading manifest file"
	}

	ext := path.Ext(sourcePath)
	switch {
	case ext == ".json":
		if err := json.Unmarshal(data, obj); err != nil {
			return fmt.Errorf("%s as JSON (%q) – %v", errorFmt, sourcePath, err)
		}
	case ext == ".yaml" || ext == ".yml":
		if err := yaml.Unmarshal(data, obj); err != nil {
			return fmt.Errorf("%s as YAML (%q) – %v", errorFmt, sourcePath, err)
		}
	case ext == ".kg" || ext == ".hcl":
		if err := NewFromHCL(obj, data); err != nil {
			return fmt.Errorf("%s as HCL (%q) – %v", errorFmt, sourcePath, err)
		}
	default:
		return fmt.Errorf("%s %q – unknown file extension", errorFmt, sourcePath)
	}

	return nil
}

func EnsureJSON(data []byte) ([]byte, error) {
	obj := new(interface{})
	if err := json.Unmarshal(data, obj); err != nil {
		return data, fmt.Errorf("invalid JSON data – %v", err)
	}
	return json.Marshal(obj)
}
