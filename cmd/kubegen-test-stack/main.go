package main

import (
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/pkg/apps"
	"github.com/errordeveloper/kubegen/pkg/util"
)

var (
	manifest     string
	outputFormat string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen-test-stack",
		RunE: command,
	}

	rootCmd.Flags().StringVarP(&manifest, "manifest", "", "Kubefile", "App manifest to use")
	rootCmd.Flags().StringVarP(&outputFormat, "output-format", "", "yaml", "App manifest to use")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	input, err := ioutil.ReadFile(manifest)
	if err != nil {
		return err
	}

	app, err := apps.NewFromHCL(input)
	if err != nil {
		return err
	}

	var data []byte
	switch outputFormat {
	case "yaml":
		if data, err = app.EncodeListToYAML(); err != nil {
			return err
		}
	case "json":
		if data, err = app.EncodeListToPrettyJSON(); err != nil {
			return err
		}
	}

	if err := util.Dump(outputFormat, data); err != nil {
		return err
	}

	return nil
}
