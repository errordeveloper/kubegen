package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/pkg/resources"
	"github.com/errordeveloper/kubegen/pkg/util"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen-test-module",
		RunE: command,
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	module, err := resources.NewModuleInstance(args[0])
	if err != nil {
		panic(err)
	}

	var data []byte
	if data, err = module.EncodeToYAML(); err != nil {
		return err
	}

	if err := util.Dump("yaml", data); err != nil {
		return err
	}

	return nil
}
