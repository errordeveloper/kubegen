package resources

import (
	"fmt"

	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ulule/deepcopier"
)

func (i Secret) ToObject(localGroup *Group) (runtime.Object, error) {
	obj, err := i.Convert(localGroup)
	if err != nil {
		return runtime.Object(nil), err
	}
	return runtime.Object(obj), nil
}

func (i *Secret) Convert(localGroup *Group) (*corev1.Secret, error) {
	meta := i.Metadata.Convert(i.Name, localGroup)

	secret := corev1.Secret{
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
			return nil, fmt.Errorf("cannot read Secret data from file %q – %v", v, err)
		}
		secret.Data[v] = data
	}

	return &secret, nil
}
