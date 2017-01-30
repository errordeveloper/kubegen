package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/appmaker"
)

var (
	flavor   string
	image    string
	manifest string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen",
		RunE: kubegen,
	}

	rootCmd.Flags().StringVarP(&manifest, "manifest", "", "", "App manifest to use")
	rootCmd.Flags().StringVarP(&image, "image", "", "", "Image to use")
	rootCmd.Flags().StringVarP(&flavor, "flavor", "", "default", "Flavor of generator to use")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func kubegen(cmd *cobra.Command, args []string) error {
	if manifest != "" {
		data, err := ioutil.ReadFile(manifest)
		if err != nil {
			return err
		}

		app, err := appmaker.NewFromHCL(data)
		if err != nil {
			return err
		}

		return encodeAndOutput(app)
	}

	if image == "" {
		return fmt.Errorf("Image must be specified")
	}

	app := &appmaker.App{
		GroupName: "test",
		Components: []appmaker.AppComponent{
			{
				Image:            image,
				AppComponentBase: appmaker.AppComponentBase{Flavor: flavor},
			},
		},
	}

	return encodeAndOutput(app)
}

func encodeAndOutput(app *appmaker.App) error {
	data, err := app.EncodeListToYAML()
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
