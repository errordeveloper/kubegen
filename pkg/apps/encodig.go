package apps

import "github.com/errordeveloper/kubegen/pkg/util"

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
