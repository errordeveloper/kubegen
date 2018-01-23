package modules

import (
	"fmt"

	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/errordeveloper/kubegen/pkg/macroproc"
	"github.com/errordeveloper/kubegen/pkg/resources"
	"github.com/errordeveloper/kubegen/pkg/util"
)

func (i *Module) makeLookupModifier(c *macroproc.Converter, branch *macroproc.BranchLocator, _ *macroproc.Macro) (macroproc.ModifierCallback, error) {
	cb := func(m *macroproc.Modifier, c *macroproc.Converter) error {

		k := m.Branch.StringValue()
		if k == nil {
			return fmt.Errorf("attribute reference is not a string – %#v", m.Branch)
		}
		v, ok := i.attributes[*k]
		if !ok {
			return fmt.Errorf("undeclared attribute %q", *k)
		}
		if err := v.typeCheck(m.Macro); err != nil {
			return err
		}
		if m.Macro.ReturnType == macroproc.Array || m.Macro.ReturnType == macroproc.Object {
			if err := c.Overlay(m.Branch, v.Value); err != nil {
				return err
			}
		} else {
			if err := c.Set(m.Branch, v.Value); err != nil {
				return err
			}
		}
		return nil
	}
	return c.TypeCheckModifier(branch, macroproc.String, cb)
}

func loadObjWithModuleContext(group *resources.Group, data []byte, sourcePath string, instanceName string, moduleContext *Module) error {
	mp := macroproc.New()

	mp.DefineMacro(macroproc.MacroStringLookup, moduleContext.makeLookupModifier)
	mp.DefineMacro(macroproc.MacroNumberLookup, moduleContext.makeLookupModifier)
	mp.DefineMacro(macroproc.MacroObjectLookup, moduleContext.makeLookupModifier)
	mp.DefineMacro(macroproc.MacroArrayLookup, moduleContext.makeLookupModifier)

	mp.DefineMacro(macroproc.MacroStringJoin, macroproc.MakeModifierStringJoin)
	mp.DefineMacro(macroproc.MacroStringAsJSON, macroproc.MakeModifierStringAsJSON)
	mp.DefineMacro(macroproc.MacroStringAsYAML, macroproc.MakeModifierStringAsYAML)

	if err := mp.LoadObject(data, sourcePath, instanceName); err != nil {
		return err
	}
	if err := mp.Run(); err != nil {
		return err
	}
	if err := mp.UnloadObject(group, sourcePath, instanceName); err != nil {
		return err
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

	if err := util.LoadObj(b, data, bundlePath, ""); err != nil {
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

		if err := m.LoadAttributes(i); err != nil {
			return err
		}

		b.loadedModules = append(b.loadedModules, *m)
		//log.Printf("Added module with %d manifests", len(m.manifests))
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
		if err := util.LoadObj(m, data, manifestPath, instanceName); err != nil {
			return nil, err
		}
		if m.Kind != ModuleKind {
			return nil, fmt.Errorf(
				"error loading file %q in module %q – unrecognised `Kind: %q`, must be %q",
				file.Name(), dir, m.Kind, ModuleKind)
		}
		// Parameters and Internals are scoped globally, here we collect them
		module.Parameters = append(module.Parameters, m.Parameters...)
		module.Internals = append(module.Internals, m.Internals...)
		// The module itself isn't something we can parse 100% yet, so we only store it as a string
		module.manifests[manifestPath] = data
	}

	return module, nil
}

func (i *ModuleParameter) load(m *Module, instance ModuleInstance) error {
	undefinedNonOptionalParameterError := fmt.Errorf(
		"module %q must set parameter %q (of type %s)",
		instance.Name, i.Name, i.Type)

	unknownParameterTypeError := fmt.Errorf(
		"parameter %q in module %q of unknown type %q, only types \"String\" and \"Number\" are supported",
		i.Name, instance.Name, i.Type)

	wrongParameterTypeError := func(v interface{}) error {
		return fmt.Errorf(
			"parameter %q in module %q not of type %q [value: %#v]",
			i.Name, instance.Name, i.Type, v)
	}

	defaultValueNotSetError := fmt.Errorf(
		"parameter %q in module %q of type %q must either be required or provide a default value",
		i.Name, instance.Name, i.Type)

	if v, ok := m.attributes[i.Name]; ok {
		return fmt.Errorf("cannot declare parameter %q in module %q (attribute already defined as %q), already defined",
			i.Name, instance.Name, v.Kind)
	}

	switch i.Type {
	case "Number":
		// all numeric values from YAML are parsed as float64, but Kubernetes API mostly wants int32
		var value int32
		v, isSet := instance.Parameters[i.Name]
		// TODO how can we safely detect if default value is set and derive whether this is optional or not from that?
		if i.Required {
			if isSet {
				switch v.(type) {
				case float64:
					value = int32(v.(float64))
				case int:
					value = int32(v.(int))
				default:
					return wrongParameterTypeError(v)
				}
			} else {
				return undefinedNonOptionalParameterError
			}
		} else {
			if isSet {
				switch v.(type) {
				case float64:
					value = int32(v.(float64))
				case int:
					value = int32(v.(int))
				default:
					return wrongParameterTypeError(v)
				}
			} else {
				if i.Default == nil {
					return defaultValueNotSetError
				}
				switch i.Default.(type) {
				case string:
					value = int32(i.Default.(float64))
				default:
					return wrongParameterTypeError(v)
				}
			}
		}
		m.attributes[i.Name] = attribute{
			Type:  i.Type,
			Value: value,
			Kind:  "parameter",
		}
		return nil
	case "String":
		var value string
		v, isSet := instance.Parameters[i.Name]
		if i.Required {
			if isSet {
				switch v.(type) {
				case string:
					value = v.(string)
				default:
					return wrongParameterTypeError(v)
				}
			} else {
				return undefinedNonOptionalParameterError
			}
		} else {
			// TODO warn if we see an empty string here as it is most likely an issue...
			if isSet {
				switch v.(type) {
				case string:
					value = v.(string)
				default:
					return wrongParameterTypeError(v)
				}
			} else {
				if i.Default == nil {
					return defaultValueNotSetError
				}
				switch i.Default.(type) {
				case string:
					value = i.Default.(string)
				default:
					return wrongParameterTypeError(v)
				}
			}
		}
		m.attributes[i.Name] = attribute{
			Type:  i.Type,
			Value: value,
			Kind:  "parameter",
		}
		return nil
	default:
		return unknownParameterTypeError
	}
}

func (i *ModuleInternal) load(m *Module, instance ModuleInstance) error {
	if v, ok := m.attributes[i.Name]; ok {
		return fmt.Errorf("cannot declare internal %q in module %q (attribute already defined as %q), already defined",
			i.Name, instance.Name, v.Kind)
	}

	m.attributes[i.Name] = attribute{
		Type:  i.Type,
		Value: i.Value,
		Kind:  "internal",
	}

	return nil
}

// TODO maybe this should be a more generic thing in pkg/macroproc, e.g. macro.TypeCheck(interface{})
func (i *attribute) typeCheck(macro *macroproc.Macro) error {
	rt := strings.Title(macro.ReturnType.String())
	cannotConvertError := fmt.Errorf("cannot convert type %s to %s for %q", i.Type, rt, macro)
	if i.Type != rt {
		return cannotConvertError
	}

	switch macro.ReturnType {
	case macroproc.Number:
		switch i.Value.(type) {
		case int:
			break
		case float64:
			break
		default:
			return cannotConvertError
		}
	case macroproc.String:
		switch i.Value.(type) {
		case string:
			break
		default:
			return cannotConvertError
		}
	}

	return nil
}

func (m *Module) LoadAttributes(instance ModuleInstance) error {
	m.attributes = make(map[string]attribute, len(m.Parameters))

	for _, parameter := range m.Parameters {
		if err := parameter.load(m, instance); err != nil {
			return err
		}
	}

	for _, internal := range m.Internals {
		if err := internal.load(m, instance); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) LoadGroups(instanceName, namespace string) (map[string]resources.Group, error) {
	groups := make(map[string]resources.Group)

	for manifestPath, data := range m.manifests {
		// TODO also do something about multiple formats here
		group := resources.Group{}
		if err := loadObjWithModuleContext(&group, data, manifestPath, instanceName, m); err != nil {
			return nil, err
		}

		// Local namespace overrides global namespace if set
		if group.Namespace == "" && namespace != "" {
			group.Namespace = namespace
		}

		groups[manifestPath] = group
		//log.Printf("Loaded group from %q", manifestPath)
		//log.Printf("groups[%s]=%#v", manifestPath, group)
	}

	return groups, nil
}

func (m *Module) EncodeGroupsToYAML(instance ModuleInstance) (map[string][]byte, error) {
	output := make(map[string][]byte)
	groups, err := m.LoadGroups(instance.Name, instance.Namespace)
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
	groups, err := m.LoadGroups(instance.Name, instance.Namespace)
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
