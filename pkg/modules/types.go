package modules

const (
	ModuleKind = "kubegen.k8s.io/Module.v1alpha2"
	BundleKind = "kubegen.k8s.io/Bundle.v1alpha2"
)

type Bundle struct {
	Kind          string           `yaml:"Kind" json:"Kind" hcl:"kind"`
	Name          string           `yaml:"Name" json:"Name" hcl:"name"`
	Namespace     string           `yaml:"Namespace,omitempty" json:"Namespace,omitempty" hcl:"namespace"`
	Description   string           `yaml:"Description,omitempty" json:"Description" hcl:"description"`
	Modules       []ModuleInstance `yaml:"Modules" "json:"Modules" hcl:"module"`
	path          string           `yaml:"-" json:"-" hcl:"-"`
	loadedModules []Module         `yaml:"-" json:"-" hcl:"-"`
}

type ModuleInstance struct {
	Name       string                 `yaml:"Name" json:"Name" hcl:",key"`
	Namespace  string                 `yaml:"Namespace,omitempty" json:"Namespace,omitempty" hcl:"namespace"`
	SourceDir  string                 `yaml:"SourceDir" json:"SourceDir" hcl:"source_dir"`
	OutputDir  string                 `yaml:"OutputDir" json:"OutputDir" hcl:"output_dir"`
	Parameters map[string]interface{} `yaml:"Parameters,omitempty" json:"Parameters,omitempty" hcl:"parameters"`
	Internals  map[string]interface{} `yaml:"Internals,omitempty" json:"Internals,omitempty" hcl:"internals"`
}

type valueLookupFunc func() []byte

type Module struct {
	Kind             string               `yaml:"Kind" json:"Kind" hcl:"kind"`
	Parameters       []ModuleParameter    `yaml:"Parameters,omitempty" json:"Parameters,omitempty" hcl:"parameter"`
	Internals        []ModuleInternal     `yaml:"Internals,omitempty" json:"Internals,omitempty" hcl:"internals"`
	manifests        map[string][]byte    `yaml:"-" json:"-" hcl:"-"`
	path             string               `yaml:"-" json:"-" hcl:"-"`
	attributes       map[string]attribute `yaml:"-" json:"-" hcl:"-"`
	IncludeManifests []RawResource        `yaml:"Include" json:"Include" hcl:"include"`
}

// TODO conditionally laod "raw" files
type RawResource struct {
	Path             string `yaml:"Path" json:"Path" hcl:",key"`
	ControlParameter string `yaml:"ControlParameter" json:"ControlParameter" hcl:"control_parameter"`
}

type ModuleParameter struct {
	Name     string      `yaml:"name" json:"name" hcl:",key"`
	Type     string      `yaml:"type" json:"type" hcl:"type"`
	Required bool        `yaml:"required" json:"required" hcl:"required"`
	Default  interface{} `yaml:"default" json:"default" hcl:"default"`
}

type ModuleInternal struct {
	Name  string      `yaml:"name" json:"name" hcl:",key"`
	Type  string      `yaml:"type" json:"type" hcl:"type"`
	Value interface{} `yaml:"default" json:"default" hcl:"default"`
}

type attribute struct {
	Type  string      `yaml:"type" json:"type" hcl:"type"`
	Value interface{} `yaml:"value" json:"value" hcl:"value"`
	Kind  string      `yaml:"kind" json:"kind" hcl:"kind"`
}
