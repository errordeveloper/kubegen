package main // import "github.com/errordeveloper/kubegen/cmd/kubegen"

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/errordeveloper/kubegen/pkg/modules"
	"github.com/errordeveloper/kubegen/pkg/util"
)

var (
	module    modules.ModuleInstance
	variables []string
)

const (
	defaultName      = "$(basename <source-dir>)"
	defaultOutputDir = "./<name>"
)

var moduleCmd = &cobra.Command{
	Use:   "module <moduleSourceDir>",
	Short: "Generate single module as specified with flags",
	RunE:  moduleFn,
}

func init() {
	//moduleCmd.Flags().StringVarP(&module.SourceDir, "source-dir", "D", "",
	//	"Module source directory (must be specified, if it is same as current working directory `--stdout` will be set)")
	moduleCmd.Flags().StringVarP(&module.OutputDir, "output-dir", "O", defaultOutputDir,
		"Output directory")

	moduleCmd.Flags().StringVarP(&module.Name, "name", "n", defaultName,
		"Name of the module instance (optional)")
	moduleCmd.Flags().StringVarP(&module.Namespace, "namespace", "N", "",
		"Namespace of the module instance (optional)")

	moduleCmd.Flags().StringSliceVarP(&variables, "variables", "v", []string{},
		"Variables to set for the module instance")
}

func moduleFn(cmd *cobra.Command, args []string) error {
	//if module.SourceDir == "" {
	if len(args) == 0 {
		return fmt.Errorf("please provide module source directory")
	}
	if len(args) > 1 {
		return fmt.Errorf("only one module source directory needed")
	}

	module.SourceDir = args[0]

	if module.Name == defaultName {
		sourceDirAbsPath, err := filepath.Abs(module.SourceDir)
		if err != nil {
			return err
		}

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		if sourceDirAbsPath == wd {
			stdout = true
		}

		module.Name = path.Base(module.SourceDir)
	}

	if module.OutputDir == defaultOutputDir {
		module.OutputDir = module.Name
	}

	if len(variables) > 0 {
		module.Variables = make(map[string]interface{})
	}

	for _, v := range variables {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) < 2 {
			return fmt.Errorf("invalid variable declaration %q, expected \"key=value\"", v)
		}

		if kv[1] == "" {
			return fmt.Errorf("invalid variable value %q, expected a non-empty string", v)
		}

		v := intstr.Parse(kv[1])
		switch v.Type {
		case intstr.Int:
			module.Variables[kv[0]] = v.IntValue()
		case intstr.String:
			module.Variables[kv[0]] = v.String()
		}
	}

	bundle := &modules.Bundle{Modules: []modules.ModuleInstance{module}}

	if err := bundle.LoadModules(nil); err != nil {
		return err
	}

	if !stdout {
		wroteFiles, err := bundle.WriteToOutputDir(format)
		if err != nil {
			return err
		}

		fmt.Printf("Wrote %d files", len(wroteFiles))

		if len(wroteFiles) == 0 {
			fmt.Printf(".\n")
			return nil
		}

		fmt.Printf(":\n")
		for _, file := range wroteFiles {
			fmt.Printf("  – %s\n", file)
		}
		return nil
	} else {
		var (
			data []byte
			err  error
		)
		switch format {
		case "yaml":
			if data, err = bundle.EncodeAllToYAML(); err != nil {
				return err
			}
		case "json":
			if data, err = bundle.EncodeAllToJSON(); err != nil {
				return err
			}
		}

		if err := util.Dump(format, data); err != nil {
			return err
		}
	}

	return nil
}
