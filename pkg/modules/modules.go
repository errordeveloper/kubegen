package modules

import (
	"fmt"

	"bytes"
	"io/ioutil"
	"os"
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

const (
	bundleReadErrorFmt        = "kubegen/modules: error reading resource group definition file %q – %v"
	resourceGroupReadErrorFmt = "kubegen/modules: error reading bundle definition file %q – %v"
	moduleReadErrorFmt        = "kubegen/modules: error reading file %q in module %q – %v"
)

func NewBundle(bundlePath string) (*Bundle, error) {
	b := &Bundle{path: bundlePath}

	data, err := ioutil.ReadFile(bundlePath)
	if err != nil {
		return nil, fmt.Errorf(bundleReadErrorFmt, bundlePath, err)
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

func (b *Bundle) WriteToOutputDir(contentType string) ([]string, error) {
	filesWritten := []string{}

	for n, i := range b.loadedModules {
		var (
			groups map[string][]byte
			err    error
		)
		switch contentType {
		case "yaml":
			groups, err = i.EncodeGroupsToYAML(b.Modules[n])
		case "json":
			groups, err = i.EncodeGroupsToJSON()
		}
		if err != nil {
			return nil, err
		}

		dir := b.Modules[n].OutputDir

		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("kubegen/modules: error creating output directory %q – %v", dir, err)
		}

		for manifestPath, group := range groups {
			outputFilename := path.Join(dir, strings.TrimSuffix(path.Base(manifestPath), path.Ext(manifestPath))+"."+contentType)
			if err := ioutil.WriteFile(outputFilename, group, 0644); err != nil {
				return nil, fmt.Errorf("kubegen/modules: error writing to file %q – %v", outputFilename, err)
			}
			filesWritten = append(filesWritten, outputFilename)
		}
	}

	return filesWritten, nil
}

func (b *Bundle) EncodeAllToYAML() ([]byte, error) {
	output := []byte{}

	for n, i := range b.loadedModules {
		groups, err := i.EncodeGroupsToYAML(b.Modules[n])
		if err != nil {
			return nil, err
		}

		for _, group := range groups {
			output = append(output, group...)
		}
	}

	return output, nil
}

func (b *Bundle) EncodeAllToJSON() ([]byte, error) {
	// TODO figure out how are we gonna merge lists
	return nil, nil
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
		m := &Module{}
		manifestPath := path.Join(dir, file.Name())
		data, err := ioutil.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf(moduleReadErrorFmt, file.Name(), dir, err)
		}
		ext := path.Ext(manifestPath)
		if ext == ".yaml" || ext == ".yml" {
			if err := yaml.Unmarshal(data, m); err != nil {
				return nil, err
			}
		}
		if ext == ".kg" || ext == ".hcl" {
			if err := util.NewFromHCL(m, data); err != nil {
				return nil, err
			}
		}
		module.Variables = append(module.Variables, m.Variables...)
		module.manifests[manifestPath] = data
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
		group, err := NewResourceGroup(manifestPath, data)
		if err != nil {
			return nil, err
		}
		groups[manifestPath] = group
	}

	return groups, nil
}

func (m *Module) EncodeGroupsToYAML(instance ModuleInstance) (map[string][]byte, error) {
	output := make(map[string][]byte)
	groups, err := m.MakeGroups()
	if err != nil {
		return nil, err
	}

	for manifestPath, group := range groups {
		data, err := group.EncodeListToYAML()
		if err != nil {
			return nil, err
		}

		info := fmt.Sprintf(
			"\n# Generated from module\n#\tName: %q\n#\tSourceDir: %q\n#\tmanifestPath: %q\n",
			instance.Name,
			instance.SourceDir,
			manifestPath,
		)

		output[manifestPath] = append([]byte(info), data...)
	}

	return output, nil
}

func (m *Module) EncodeGroupsToJSON() (map[string][]byte, error) {
	output := make(map[string][]byte)
	_, err := m.MakeGroups()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func NewResourceGroup(groupPath string, data []byte) (*resources.Group, error) {
	group := &resources.Group{}
	// TODO we do not load vanilla YAML or JSON directly here, we can handle those via
	// an explicit declaration
	ext := path.Ext(groupPath)
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, group); err != nil {
			return nil, fmt.Errorf("kubegen/resources: error converting YAML to internal representation – %v", err)
		}
		return group, nil
	}
	if ext == ".kg" || ext == ".hcl" {
		if err := util.NewFromHCL(group, data); err != nil {
			return nil, err
		}
		return group, nil
	}

	return nil, fmt.Errorf(resourceGroupReadErrorFmt, groupPath, "unknown file extention")
}
