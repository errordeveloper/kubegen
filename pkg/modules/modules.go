package modules

import (
	"fmt"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/errordeveloper/kubegen/pkg/resources"
	"github.com/errordeveloper/kubegen/pkg/util"
)

// TODO report unknown keys in manifests
// TODO bail on unknown variable keys

func loadFromPath(obj interface{}, data []byte, sourcePath string, instanceName string) error {
	var errorFmt string
	if instanceName != "" {
		errorFmt = fmt.Sprintf("error loading module %q source", instanceName)
	} else {
		errorFmt = "error loading manifest file"
	}

	ext := path.Ext(sourcePath)
	switch {
	case ext == ".json":
		if err := json.Unmarshal(data, obj); err != nil {
			return fmt.Errorf("%s as JSON (%q) – %v", errorFmt, sourcePath, err)
		}
	case ext == ".yaml" || ext == ".yml":
		if err := yaml.Unmarshal(data, obj); err != nil {
			return fmt.Errorf("%s as YAML (%q) – %v", errorFmt, sourcePath, err)
		}
	case ext == ".kg" || ext == ".hcl":
		if err := util.NewFromHCL(obj, data); err != nil {
			return fmt.Errorf("%s as HCL (%q) – %v", errorFmt, sourcePath, err)
		}
	default:
		return fmt.Errorf("%s %q – unknown file extension", errorFmt, sourcePath)
	}
	return nil
}

func NewBundle(bundlePath string) (*Bundle, error) {
	b := &Bundle{path: bundlePath}

	data, err := ioutil.ReadFile(bundlePath)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading bundle manifest file %q – %v",
			bundlePath, err)
	}

	if err := loadFromPath(b, data, bundlePath, ""); err != nil {
		return nil, err
	}
	if b.Kind != BundleKind {
		return nil, fmt.Errorf(
			"error loading bundle manifest %q – unrecognised `Kind: %q`, must be %q",
			bundlePath, b.Kind, BundleKind)
	}

	return b, nil
}

func (b *Bundle) LoadModules(selectNames []string) error {
	applyNameSelector := len(selectNames) > 0

	for n, i := range b.Modules {
		if applyNameSelector {
			skip := true
			for _, name := range selectNames {
				if i.Name == name {
					skip = false
				}
			}
			if skip {
				continue
			}
		}

		var err error
		m, err := NewModule(path.Join(path.Dir(b.path), i.SourceDir), i.Name)
		if err != nil {
			return err
		}

		// Local namespace overrides global namespace if set
		if i.Namespace == "" && b.Namespace != "" {
			b.Modules[n].Namespace = b.Namespace
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
			groups, err = i.EncodeGroupsToJSON(b.Modules[n])
		}
		if err != nil {
			return nil, err
		}

		dir := b.Modules[n].OutputDir

		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("error creating output directory %q – %v", dir, err)
		}

		for manifestPath, group := range groups {
			outputFilename := path.Join(dir, strings.TrimSuffix(path.Base(manifestPath), path.Ext(manifestPath))+"."+contentType)
			if err := ioutil.WriteFile(outputFilename, group, 0644); err != nil {
				return nil, fmt.Errorf("error writing to file %q – %v", outputFilename, err)
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

		sortedGroups := []string{}
		for manifestPath, _ := range groups {
			sortedGroups = append(sortedGroups, manifestPath)
		}

		sort.Strings(sortedGroups)

		for _, manifestPath := range sortedGroups {
			output = append(output, groups[manifestPath]...)
		}
	}

	return output, nil
}

func (b *Bundle) EncodeAllToJSON() ([]byte, error) {
	output := []byte{}

	for n, i := range b.loadedModules {
		groups, err := i.EncodeGroupsToJSON(b.Modules[n])
		if err != nil {
			return nil, err
		}

		sortedGroups := []string{}
		for manifestPath, _ := range groups {
			sortedGroups = append(sortedGroups, manifestPath)
		}

		sort.Strings(sortedGroups)

		for _, manifestPath := range sortedGroups {
			output = append(output, groups[manifestPath]...)
		}
	}

	return output, nil
}

func NewModule(dir, instanceName string) (*Module, error) {
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
			return nil, fmt.Errorf(
				"error reading file %q in module %q – %v",
				file.Name(), dir, err)
		}
		if err := loadFromPath(m, data, manifestPath, instanceName); err != nil {
			return nil, err
		}
		if m.Kind != ModuleKind {
			return nil, fmt.Errorf(
				"error loading file %q in module %q – unrecognised `Kind: %q`, must be %q",
				file.Name(), dir, m.Kind, ModuleKind)
		}
		// Variables and partials are scoped globally, here we collect them
		module.Variables = append(module.Variables, m.Variables...)
		module.Partials = append(module.Partials, m.Partials...)
		// The module itself isn't something we can parse 100% yet, so we only store it as a string
		module.manifests[manifestPath] = data
	}

	return module, nil
}

func (i *ModuleVariable) makeValFunc(instance ModuleInstance) (interface{}, error) {
	undefinedNonOptionalVariableError := fmt.Errorf(
		"module %q must set variable %q (of type %s)",
		instance.Name, i.Name, i.Type)

	unknownVariableTypeError := fmt.Errorf(
		"variable %q in module %q of unknown type %q, only types \"string\" and \"number\" are supported",
		i.Name, instance.Name, i.Type)

	wrongVariableTypeError := func(v interface{}) error {
		return fmt.Errorf(
			"variable %q in module %q not of type %q [value: %#v]",
			i.Name, instance.Name, i.Type, v)
	}

	defaultValueNotSetError := fmt.Errorf(
		"variable %q in module %q of type %q must either be required or provide a default value",
		i.Name, instance.Name, i.Type)

	switch i.Type {
	case "number":
		// all numeric values from YAML are parsed as float64, but Kubernetes API mostly wants int32
		var value int32
		v, isSet := instance.Variables[i.Name]
		// TODO how can we safely detect if default value is set and derive whether this is optional or not from that?
		if i.Required {
			if isSet {
				switch v.(type) {
				case float64:
					value = int32(v.(float64))
				case int:
					value = int32(v.(int))
				default:
					return nil, wrongVariableTypeError(v)
				}
			} else {
				return nil, undefinedNonOptionalVariableError
			}
		} else {
			if isSet {
				switch v.(type) {
				case float64:
					value = int32(v.(float64))
				case int:
					value = int32(v.(int))
				default:
					return nil, wrongVariableTypeError(v)
				}
			} else {
				if i.Default == nil {
					return nil, defaultValueNotSetError
				}
				switch i.Default.(type) {
				case string:
					value = int32(i.Default.(float64))
				default:
					return nil, wrongVariableTypeError(v)
				}
			}
		}
		return func() int32 { return value }, nil
	case "string":
		var value string
		v, isSet := instance.Variables[i.Name]
		if i.Required {
			if isSet {
				switch v.(type) {
				case string:
					value = v.(string)
				default:
					return nil, wrongVariableTypeError(v)
				}
			} else {
				return nil, undefinedNonOptionalVariableError
			}
		} else {
			// TODO warn if we see an empty string here as it is most likely an issue...
			if isSet {
				switch v.(type) {
				case string:
					value = v.(string)
				default:
					return nil, wrongVariableTypeError(v)
				}
			} else {
				if i.Default == nil {
					return nil, defaultValueNotSetError
				}
				switch i.Default.(type) {
				case string:
					value = i.Default.(string)
				default:
					return nil, wrongVariableTypeError(v)
				}
			}
		}
		return func() string { return value }, nil
	default:
		return nil, unknownVariableTypeError
	}
}

func (m *Module) Load(instance ModuleInstance) error {
	// TODO we shuld probably use, as has sane and widely-used syntax,
	// it is also fairly restrictive https://github.com/hoisie/mustache
	funcMap := template.FuncMap{}
	for _, variable := range m.Variables {
		valFunc, err := variable.makeValFunc(instance)
		if err != nil {
			return err
		}

		funcMap[variable.Name] = valFunc
	}

	for manifestPath, data := range m.manifests {
		output := bytes.Buffer{}
		// Let's use some very familiar delimitors
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

func (m *Module) MakeGroups(instanceName, namespace string) (map[string]resources.Group, error) {
	groups := make(map[string]resources.Group)

	for manifestPath, data := range m.manifests {
		// TODO also do something about multiple formats here
		group := resources.Group{}
		if err := loadFromPath(&group, data, manifestPath, instanceName); err != nil {
			return nil, err
		}

		// Local namespace overrides global namespace if set
		if group.Namespace == "" && namespace != "" {
			group.Namespace = namespace
		}

		groups[manifestPath] = group
	}

	return groups, nil
}

func (m *Module) EncodeGroupsToYAML(instance ModuleInstance) (map[string][]byte, error) {
	output := make(map[string][]byte)
	groups, err := m.MakeGroups(instance.Name, instance.Namespace)
	if err != nil {
		return nil, err
	}

	for manifestPath, group := range groups {
		data, err := group.EncodeListToYAML()
		if err != nil {
			return nil, err
		}

		if data == nil {
			continue
		}

		info := fmt.Sprintf(
			"\n---\n#\n# Generated from module\n#\tName: %q\n#\tSourceDir: %q\n#\tmanifestPath: %q\n#\n\n",
			instance.Name,
			instance.SourceDir,
			manifestPath,
		)

		output[manifestPath] = append([]byte(info), data...)
	}

	return output, nil
}

func (m *Module) EncodeGroupsToJSON(instance ModuleInstance) (map[string][]byte, error) {
	output := make(map[string][]byte)
	groups, err := m.MakeGroups(instance.Name, instance.Namespace)
	if err != nil {
		return nil, err
	}

	for manifestPath, group := range groups {
		data, err := group.EncodeListToPrettyJSON()
		if err != nil {
			return nil, err
		}

		if data == nil {
			continue
		}

		output[manifestPath] = append(data, byte('\n'))
	}

	return output, nil
}
