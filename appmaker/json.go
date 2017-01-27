package appmaker

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	kapi "k8s.io/client-go/pkg/api/v1"
)

func marshalMultipleToJSON(resources map[int]interface{}) (map[int][]byte, error) {
	var err error

	keys := []int{}
	for k, _ := range resources {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	data := map[int][]byte{}
	for _, j := range keys {
		for k, v := range resources {
			if k == j {
				if data[k], err = json.Marshal(v); err != nil {
					return nil, err
				}
			}
		}
	}
	return data, nil
}

func NewFromJSON(manifest []byte) (*App, error) {
	app := &App{}
	if err := json.Unmarshal(manifest, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (i *App) MarshalToCombinedJSON() ([]byte, error) {
	data, err := i.MarshalToJSON()
	if err != nil {
		return nil, err
	}

	fragments := []string{}
	for _, component := range data {
		for _, resource := range component {
			fragments = append(fragments, string(resource))
		}
	}

	listFormat := `{"apiVersion":"v1","kind":"List","items":[%s]}`
	return []byte(fmt.Sprintf(listFormat, strings.Join(fragments, ","))), nil
}

func (i *App) ToList() (*kapi.List, error) {
	data, err := i.MarshalToCombinedJSON()
	if err != nil {
		return nil, err
	}

	list := &kapi.List{}
	if err := json.Unmarshal(data, list); err != nil {
		return nil, err
	}

	return list, nil
}

func (i *App) MarshalIndentToCombinedJSON() ([]byte, error) {
	list, err := i.ToList()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(list, "", "  ")
}

func (i *App) MarshalToJSON() ([]map[int][]byte, error) {
	components := i.MakeAll()
	data := make([]map[int][]byte, len(components))

	keys := []int{}
	for k, _ := range components {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	for _, j := range keys {
		for k, v := range components {
			if k == j {
				componentData, err := v.MarshalToJSON()
				if err != nil {
					return nil, err
				}
				data = append(data, componentData)
			}
		}
	}
	return data, nil
}

func (i *AppComponent) MarshalToJSON(params AppParams) (map[int][]byte, error) {
	return i.MakeAll(params).MarshalToJSON()
}

func (i *AppComponentResources) MarshalToJSON() (map[int][]byte, error) {
	resources := make(map[int]interface{})

	switch i.manifest.Kind {
	case Deployment:
		resources[Deployment] = i.deployment
	}

	if i.service != nil {
		resources[Service] = i.service
	}

	//if i.configMap != nil { ...
	//if i.secret != nil { ...

	data, err := marshalMultipleToJSON(resources)
	if err != nil {
		return nil, err
	}
	return data, nil
}
