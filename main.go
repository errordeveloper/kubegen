package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/appmaker"
)

var (
	flavor string
	image  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen",
		RunE: kubegen,
	}

	rootCmd.Flags().StringVarP(&image, "image", "", "", "Flavor of generator to use")
	rootCmd.Flags().StringVarP(&flavor, "flavor", "", "default", "Flavor of generator to use")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func kubegen(cmd *cobra.Command, args []string) error {
	if image == "" {
		return fmt.Errorf("Image must be specified")
	}

	app := appmaker.App{
		GroupName: "test",
		Components: []appmaker.AppComponent{
			{
				Flavor: flavor,
				Image:  image,
			},
		},
	}

	data, err := app.EncodeListToYAML()
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
