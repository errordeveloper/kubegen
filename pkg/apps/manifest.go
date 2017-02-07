package apps

import (
	"encoding/json"
	"github.com/errordeveloper/kubegen/pkg/util"
)

func NewFromHCL(data []byte) (*App, error) {
	app := &App{}

	if err := util.NewFromHCL(app, data); err != nil {
		return nil, err
	}

	return app, nil
}

func (i *App) EncodeListToYAML() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return []byte{}, err
	}
	return util.EncodeList(list, "application/yaml", false)
}

func (i *App) EncodeListToJSON() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return []byte{}, err
	}
	return util.EncodeList(list, "application/json", false)
}

func (i *App) EncodeListToPrettyJSON() ([]byte, error) {
	list, err := i.MakeList()
	if err != nil {
		return []byte{}, err
	}
	return util.EncodeList(list, "application/json", true)
}

func (i *App) DumpListToFilesAsYAML() ([]string, error) {
	list, err := i.MakeList()
	if err != nil {
		return []string{}, err
	}
	return util.DumpListToFiles(list, "application/yaml")
}

// TODO we do not need most of this, but removing it will require rewriting the unit test

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

	components, err := i.MakeAll()
	if err != nil {
		return nil, err
	}

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
