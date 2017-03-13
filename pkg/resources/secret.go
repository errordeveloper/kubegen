package resources

import (
	"fmt"

	"encoding/json"
	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/ghodss/yaml"
	"github.com/ulule/deepcopier"
)

func (i Secret) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *Secret) Convert(localGroup *Group) (*v1.Secret, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
	}

	deepcopier.Copy(i).To(&secret)

	if len(secret.Data) == 0 {
		secret.Data = make(map[string][]byte)
	}

	for _, v := range i.ReadFromFiles {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, fmt.Errorf("cannot read Secret %q data from file %q – %v", v, err)
		}
		secret.Data[v] = data
	}

	for k, v := range i.EncodeAsJSON {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Secret %q data to JSON – %v", v, err)
		}
		secret.Data[k] = data
	}

	for k, v := range i.EncodeAsYAML {
		data, err := yaml.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Secret %q data to YAML – %v", v, err)
		}
		secret.Data[k] = data
	}

	return &secret, nil
}
