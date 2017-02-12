package modules

type Bundle struct {
	Name          string           `yaml:"Name" json:"Name" hcl:",key"`
	Description   string           `yaml:"Description,omitempty" json:"Description" hcl:"description"`
	Modules       []ModuleInstance `yaml:"Modules" "json:"Modules" hcl:"module"`
	path          string           `yaml:"-" json:"-" hcl:"-"`
	loadedModules []Module         `yaml:"-" json:"-" hcl:"-"`
}

type ModuleInstance struct {
	Name      string                 `yaml:"Name,omitempty" json:"Name" hcl:",key"`
	SourceDir string                 `yaml:"SourceDir" json:"SourceDir" hcl:"source_dir"`
	OutputDir string                 `yaml:"OutputDir" json:"OutputDir" hcl:"output_dir"`
	Variables map[string]interface{} `yaml:"Variables" json:"Variables" hcl:"variables"`
}

type Module struct {
	Variables        []ModuleVariable  `yaml:"Variables,omitempty" json:"Variables" hcl:"variable"`
	manifests        map[string][]byte `yaml:"-" json:"-" hcl:"-"`
	path             string            `yaml:"-" json:"-" hcl:"-"`
	IncludeManifests []RawResource     `yaml:"Include" json:"Include" hcl:"include"`
}

// TODO conditionally laod "raw" files
type RawResource struct {
	Path            string `yaml:"Path" json:"Path" hcl:",key"`
	ControlVariable string `yaml:"ControlVariable" json:"ControlVariable" hcl:"control_variable"`
}

type ModuleVariable struct {
	Name     string      `yaml:"name" json:"name" hcl:",key"`
	Type     string      `yaml:"type" json:"type" hcl:"type"`
	Required bool        `yaml:"required" json:"required" hcl:"required"`
	Default  interface{} `yaml:"default" json:"default" hcl:"default"`
}
