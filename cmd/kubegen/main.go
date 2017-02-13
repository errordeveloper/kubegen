package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/pkg/modules"
	"github.com/errordeveloper/kubegen/pkg/util"
)

var (
	stdout        bool
	format        string
	selectModules []string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen <bundleManifest> ...",
		RunE: command,
	}

	rootCmd.Flags().BoolVarP(&stdout, "stdout", "s", false,
		"Output to stdout instead of creating files")
	rootCmd.Flags().StringVarP(&format, "output", "o", "yaml",
		"Output format [\"yaml\" or \"json\"]")

	rootCmd.Flags().StringSliceVarP(&selectModules, "module", "m", []string{},
		"Names of modules to process (all modules in each given bundle are processed by defult)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide at least one manifest file")
	}

	for _, manifest := range args {
		bundle, err := modules.NewBundle(manifest)
		if err != nil {
			return err
		}

		if err := bundle.LoadModules(selectModules); err != nil {
			return err
		}

		if !stdout {
			wroteFiles, err := bundle.WriteToOutputDir(format)
			if err != nil {
				return err
			}

			fmt.Printf("Wrote %d files based on bundle manifest %q", len(wroteFiles), manifest)

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
			var data []byte
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
	}

	return nil
}
