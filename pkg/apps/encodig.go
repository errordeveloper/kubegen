package apps

import "github.com/errordeveloper/kubegen/pkg/util"

func (i *App) EncodeListToYAML() ([]byte, error) {
	return util.EncodeList(i.MakeList(), "application/yaml", false)
}

func (i *App) EncodeListToJSON() ([]byte, error) {
	return util.EncodeList(i.MakeList(), "application/json", false)
}

func (i *App) EncodeListToPrettyJSON() ([]byte, error) {
	return util.EncodeList(i.MakeList(), "application/json", true)
}

func (i *App) DumpListToFilesAsYAML() ([]string, error) {
	return util.DumpListToFiles(i.MakeList(), "application/yaml")
}
