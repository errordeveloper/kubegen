package modules

import (
	"github.com/errordeveloper/kubegen/pkg/resources"
)

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

type ManifestPath = string
type AttributeKey = string
type Module struct {
	Kind       string            `yaml:"Kind" json:"Kind" hcl:"kind"`
	Parameters []ModuleParameter `yaml:"Parameters,omitempty" json:"Parameters,omitempty" hcl:"parameter"`
	Internals  []ModuleInternal  `yaml:"Internals,omitempty" json:"Internals,omitempty" hcl:"internals"`
	Resources  []AnyResource     `yaml:"Resources" json:"Resources" hcl:"resource"`

	directory  string
	attributes map[AttributeKey]attribute
	manifests  map[ManifestPath][]byte
	resources  map[ManifestPath][]resources.Anything
}

type AnyResource struct {
	Path string `yaml:"path" json:"path" hcl:",key"`

	includedBy ManifestPath
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
	Value interface{} `yaml:"value" json:"value" hcl:"value"`
}

type attribute struct {
	Type  string      `yaml:"type" json:"type" hcl:"type"`
	Value interface{} `yaml:"value" json:"value" hcl:"value"`
	Kind  string      `yaml:"kind" json:"kind" hcl:"kind"`
}
