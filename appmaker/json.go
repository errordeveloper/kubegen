package appmaker

// TODO we do not need most of this, but removing it will require rewriting the unit test

import (
	"encoding/json"
)

func marshalMultipleToJSON(resources map[int]interface{}) (map[int][]byte, error) {
	var err error

	data := map[int][]byte{}
	for k, v := range resources {
		if data[k], err = json.Marshal(v); err != nil {
			return nil, err
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

func (i *App) MarshalToJSON() ([]map[int][]byte, error) {
	var err error

	components := i.MakeAll()

	data := make([]map[int][]byte, len(components))
	for k, v := range components {
		if data[k], err = v.MarshalToJSON(); err != nil {
			return nil, err
		}
	}
	return data, nil
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
