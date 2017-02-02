package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/pkg/apps"
	"github.com/errordeveloper/kubegen/pkg/util"
)

var (
	flavor       string
	image        string
	outputFormat string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen",
		RunE: command,
	}

	rootCmd.Flags().StringVarP(&image, "image", "", "", "Image to use")
	rootCmd.Flags().StringVarP(&flavor, "flavor", "", "default", "Flavor of generator to use")
	rootCmd.Flags().StringVarP(&outputFormat, "output-format", "", "yaml", "App manifest to use")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	if image == "" {
		return fmt.Errorf("Image must be specified")
	}

	app := &apps.App{
		GroupName: "test",
		Components: []apps.AppComponent{
			{
				Image:  image,
				Flavor: flavor,
			},
		},
	}

	var (
		data []byte
		err  error
	)
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
