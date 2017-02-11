package modules

import (
	"fmt"

	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/errordeveloper/kubegen/pkg/resources"
	"github.com/errordeveloper/kubegen/pkg/util"
)

type Bundle struct {
	Name          string           `yaml:"Name,omitempty"`
	Description   string           `yaml:"Description,omitempty"`
	Modules       []ModuleInstance `yaml:"Modules,omitempty"`
	path          string           `yaml:"-"`
	loadedModules []Module         `yaml:"-"`
}

type ModuleInstance struct {
	Name      string                 `yaml:"Name,omitempty"`
	SourceDir string                 `yaml:"SourceDir"`
	OutputDir string                 `yaml:"OutputDir"`
	Variables map[string]interface{} `yaml:"Variables"`
}

type Module struct {
	Variables []ModuleVariable  `yaml:"Variables,omitempty"`
	manifests map[string][]byte `yaml:"-"`
	path      string            `yaml:"-"`
}

type ModuleVariable struct {
	Name     string      `yaml:"name"`
	Type     string      `yaml:"type"`
	Optional bool        `yaml:"optional"`
	Default  interface{} `yaml:"default"`
}

func NewBundle(bundlePath string) (*Bundle, error) {
	b := &Bundle{path: bundlePath}

	data, err := ioutil.ReadFile(bundlePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bundle) LoadModules() error {
	for _, i := range b.Modules {
		var err error
		m, err := NewModule(path.Join(path.Dir(b.path), i.SourceDir))
		if err != nil {
			return err
		}

		if err := m.Load(i); err != nil {
			return err
		}

		b.loadedModules = append(b.loadedModules, *m)
	}

	return nil
}

func (b *Bundle) EncodeToYAML() ([]byte, error) {
	output := []byte{}

	for n, i := range b.loadedModules {
		groups, err := i.MakeGroups()
		if err != nil {
			return []byte{}, err
		}

		for manifestPath, group := range groups {
			data, err := group.EncodeListToYAML()
			if err != nil {
				return []byte{}, err
			}

			info := fmt.Sprintf(
				"\n# Generated from module\n#\tName: %q\n#\tSourceDir: %q\n#\tmanifestPath: %q\n",
				b.Modules[n].Name,
				b.Modules[n].SourceDir,
				manifestPath,
			)

			output = append(output, []byte(info)...)
			output = append(output, data...)
		}

	}

	return output, nil
}

func NewModule(dir string) (*Module, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	module := &Module{
		path:      dir,
		manifests: make(map[string][]byte),
	}
	for _, file := range files {
		// TODO consolidate with NewResourceGroupFromFile
		if strings.HasSuffix(file.Name(), "kg.yaml") || strings.HasSuffix(file.Name(), ".kg.yml") {
			m := &Module{}
			manifestPath := path.Join(dir, file.Name())
			data, err := ioutil.ReadFile(manifestPath)
			if err != nil {
				return nil, err
			}
			if err := yaml.Unmarshal(data, m); err != nil {
				return nil, err
			}
			module.Variables = append(module.Variables, m.Variables...)
			module.manifests[manifestPath] = data
		}
	}

	return module, nil
}

func (m *Module) Load(instance ModuleInstance) error {
	// TODO we shuld probably use, as has sane and widely-used syntax,
	// it is also fairly restrictive https://github.com/hoisie/mustache
	funcMap := template.FuncMap{}
	for _, variable := range m.Variables {
		undefinedNonOptionalVariableError := fmt.Errorf("module instance must set variable %q (of type %s)", variable.Name, variable.Type)
		unknownVariableTypeError := fmt.Errorf("variable %q of unknown type %q, only types \"string\" and \"number\" are supported", variable.Name, variable.Type)

		switch variable.Type {
		case "number":
			// all numeric values from YAML are parsed as float64, but Kubernetes API mostly wants int32
			var value int32
			v, isSet := instance.Variables[variable.Name]
			if variable.Optional {
				if isSet {
					value = int32(v.(float64))
				} else {
					value = int32(variable.Default.(float64))
				}
			} else {
				if isSet {
					value = int32(v.(float64))
				} else {
					return undefinedNonOptionalVariableError
				}
			}
			funcMap[variable.Name] = func() int32 { return value }
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
					return undefinedNonOptionalVariableError
				}
			}
			funcMap[variable.Name] = func() string { return value }
		default:
			return unknownVariableTypeError
		}
	}

	var output bytes.Buffer
	for manifestPath, data := range m.manifests {
		// Let's use our very familiar delimitors
		t, err := template.New(manifestPath).Delims("<", ">").Funcs(funcMap).Parse(string(data))
		if err != nil {
			return err
		}
		if err := t.Execute(&output, nil); err != nil {
			return err
		}
		m.manifests[manifestPath] = output.Bytes()
	}

	return nil
}

func (m *Module) MakeGroups() (map[string]*resources.Group, error) {
	groups := make(map[string]*resources.Group)
	for manifestPath, data := range m.manifests {
		// TODO also do something about multiple formats here
		group, err := NewResourceGroupFromYAML(data)
		if err != nil {
			return nil, err
		}
		groups[manifestPath] = group
	}

	return groups, nil
}

// TODO all of these need refactoring

func NewResourceGroupFromFile(path string) (*resources.Group, error) {
	const errfmt = "kubegen/resources: error reading resource group definition file %q – %v"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(errfmt, path, err)
	}

	var group *resources.Group
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
	// TODO allow for vanilla JSON in a module, also should allow for JSON equivalent of our YAML syntax
	//}

	return nil, fmt.Errorf(errfmt, path, "unknown file extention")
}

func NewResourceGroupFromHCL(data []byte) (*resources.Group, error) {
	group := &resources.Group{}

	if err := util.NewFromHCL(group, data); err != nil {
		return nil, err
	}

	return group, nil
}

func NewResourceGroupFromYAML(data []byte) (*resources.Group, error) {
	group := &resources.Group{}

	if err := yaml.Unmarshal(data, group); err != nil {
		return nil, fmt.Errorf("kubegen/resources: error converting YAML to internal representation – %v", err)
	}

	return group, nil
}
