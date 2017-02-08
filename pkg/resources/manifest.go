package resources

import (
	"fmt"

	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/errordeveloper/kubegen/pkg/util"
)

type Module struct {
	Variables []ModuleVariable  `yaml:"Variables"`
	files     map[string][]byte `yaml:"-"`
	path      string            `yaml:"-"`
}

type ModuleInstance struct {
	Path      string                 `yaml:"Path"`
	Variables map[string]interface{} `yaml:"Variables"`
	module    *Module                `yaml:"-"`
}

type ModuleVariable struct {
	Name     string      `yaml:"name"`
	Type     string      `yaml:"type"`
	Optional bool        `yaml:"optional"`
	Default  interface{} `yaml:"default"`
}

func NewModule(dir string) (*Module, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	module := &Module{
		path:  dir,
		files: make(map[string][]byte),
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "kg.yaml") || strings.HasSuffix(file.Name(), ".kg.yml") {
			m := &Module{}
			data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			if err := yaml.Unmarshal(data, m); err != nil {
				return nil, err
			}
			module.Variables = append(module.Variables, m.Variables...)
			module.files[file.Name()] = data
		}
	}

	return module, nil
}

func (m *Module) Load(instance *ModuleInstance) error {
	funcMap := template.FuncMap{}
	for _, variable := range m.Variables {
		switch variable.Type {
		case "number":
			var value int
			v, isSet := instance.Variables[variable.Name]
			if variable.Optional {
				if isSet {
					value = int(v.(float64))
				} else {
					value = int(variable.Default.(float64))
				}
			} else {
				if isSet {
					value = int(v.(float64))
				} else {
					return fmt.Errorf("module instance must set variable %q", variable.Name)
				}
			}
			funcMap[variable.Name] = func() int { return value }
		case "string":
			var value string
			v, isSet := instance.Variables[variable.Name]
			if variable.Optional {
				if isSet {
					value = v.(string)
				} else {
					value = variable.Default.(string)
				}
			} else {
				if isSet {
					value = v.(string)
				} else {
					return fmt.Errorf("module instance must set variable %q", variable.Name)
				}
			}
			funcMap[variable.Name] = func() string { return value }
		}
	}

	var output bytes.Buffer
	for filename, data := range m.files {
		t, err := template.New(filename).Delims("<", ">").Funcs(funcMap).Parse(string(data))
		if err != nil {
			return err
		}
		if err := t.Execute(&output, nil); err != nil {
			return err
		}
		m.files[filename] = output.Bytes()
	}

	return nil
}

func (m *Module) MakeGroups() (map[string]*ResourceGroup, error) {
	groups := make(map[string]*ResourceGroup)
	for filename, data := range m.files {
		group, err := NewResourceGroupFromYAML(data)
		if err != nil {
			return nil, err
		}
		groups[filename] = group
	}

	return groups, nil
}

func NewModuleInstance(path string) (*ModuleInstance, error) {
	i := &ModuleInstance{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, i); err != nil {
		return nil, err
	}

	m, err := NewModule(i.Path)
	if err != nil {
		return nil, err
	}

	if err := m.Load(i); err != nil {
		return nil, err
	}

	i.module = m

	return i, nil
}

func (i *ModuleInstance) EncodeToYAML() ([]byte, error) {
	output := []byte{}
	groups, err := i.module.MakeGroups()
	if err != nil {
		return []byte{}, err
	}

	for filename, group := range groups {
		data, err := group.EncodeListToYAML()
		if err != nil {
			return []byte{}, err
		}
		output = append(output, []byte(fmt.Sprintf("\n# filename: %q\n", filename))...)
		output = append(output, data...)
	}

	return output, nil
}

func NewResourceGroupFromFile(path string) (*ResourceGroup, error) {
	const errfmt = "kubegen/resources: error reading resource group definition file %q – %v"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(errfmt, path, err)
	}

	var group *ResourceGroup
	if strings.HasSuffix(path, "kg.yaml") || strings.HasSuffix(path, ".kg.yml") {
		group, err = NewResourceGroupFromYAML(data)
		if err != nil {
			return nil, err
		}
		return group, nil
	}
	//if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
	// TODO allow for vanilla YAML in a module
	//}
	if strings.HasSuffix(path, ".kg") || strings.HasSuffix(path, ".kg.hcl") || strings.HasSuffix(path, ".kg.json") {
		group, err = NewResourceGroupFromHCL(data)
		if err != nil {
			return nil, err
		}
		return group, nil
	}
	//if strings.HasSuffix(path, ".json") {
	// TODO allow for vanilla JSON in a module
	//}

	return nil, fmt.Errorf(errfmt, path, "unknown file extention")
}

func NewResourceGroupFromHCL(data []byte) (*ResourceGroup, error) {
	group := &ResourceGroup{}

	if err := util.NewFromHCL(group, data); err != nil {
		return nil, err
	}

	return group, nil
}

func NewResourceGroupFromYAML(data []byte) (*ResourceGroup, error) {
	group := &ResourceGroup{}

	if err := yaml.Unmarshal(data, group); err != nil {
		return nil, fmt.Errorf("kubegen/resources: error converting YAML to internal representation – %v", err)
	}

	return group, nil
}
